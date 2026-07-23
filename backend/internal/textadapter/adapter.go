// Package textadapter owns synchronous text-completion integration. It is
// intentionally independent from worker/adapter, which only handles media jobs.
package textadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
)

type Request struct {
	ModelName    string
	SystemPrompt string
	Prompt       string
}

type Adapter interface {
	Complete(context.Context, *Request) (string, error)
}

func GetAdapter(aiModel *model.AIModel, channel *model.AIProviderAccount) (Adapter, error) {
	if channel == nil {
		return nil, fmt.Errorf("文本模型没有可用渠道")
	}
	config := map[string]interface{}{"api_url": channel.APIURL, "api_key": channel.APIKey}
	if len(channel.ExtraConfig) > 0 {
		if err := json.Unmarshal(channel.ExtraConfig, &config); err == nil {
			config["api_url"] = channel.APIURL
			config["api_key"] = channel.APIKey
		}
	}
	switch channel.Provider {
	case "openai", "chatgpt", "grok":
		return &openAICompatible{config: config}, nil
	case "gemini":
		return &gemini{config: config}, nil
	default:
		return nil, fmt.Errorf("文本优化暂不支持渠道: %s", channel.Provider)
	}
}

type openAICompatible struct{ config map[string]interface{} }

func (a *openAICompatible) Complete(ctx context.Context, input *Request) (string, error) {
	endpoint, err := completionEndpoint(stringValue(a.config, "api_url"), stringValue(a.config, "chat_path"))
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(map[string]interface{}{
		"model": input.ModelName,
		"messages": []map[string]string{
			{"role": "system", "content": input.SystemPrompt},
			{"role": "user", "content": input.Prompt},
		},
		"temperature": 0.4,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if key := stringValue(a.config, "api_key"); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := doJSON(ctx, req, &response); err != nil {
		return "", err
	}
	if len(response.Choices) == 0 || strings.TrimSpace(response.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("文本渠道未返回优化结果")
	}
	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}

type gemini struct{ config map[string]interface{} }

func (a *gemini) Complete(ctx context.Context, input *Request) (string, error) {
	base := strings.TrimRight(stringValue(a.config, "api_url"), "/")
	if base == "" {
		return "", fmt.Errorf("渠道未配置 API 地址")
	}
	if !strings.Contains(base, "/v1beta") {
		base += "/v1beta"
	}
	endpoint := base + "/models/" + url.PathEscape(input.ModelName) + ":generateContent"
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set("key", stringValue(a.config, "api_key"))
	parsed.RawQuery = query.Encode()
	payload, err := json.Marshal(map[string]interface{}{
		"system_instruction": map[string]interface{}{"parts": []map[string]string{{"text": input.SystemPrompt}}},
		"contents":           []map[string]interface{}{{"role": "user", "parts": []map[string]string{{"text": input.Prompt}}}},
		"generationConfig":   map[string]interface{}{"temperature": 0.4},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, parsed.String(), bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := doJSON(ctx, req, &response); err != nil {
		return "", err
	}
	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 || strings.TrimSpace(response.Candidates[0].Content.Parts[0].Text) == "" {
		return "", fmt.Errorf("文本渠道未返回优化结果")
	}
	return strings.TrimSpace(response.Candidates[0].Content.Parts[0].Text), nil
}

func completionEndpoint(raw, customPath string) (string, error) {
	base := strings.TrimRight(raw, "/")
	if base == "" {
		return "", fmt.Errorf("渠道未配置 API 地址")
	}
	if customPath != "" {
		return base + "/" + strings.TrimLeft(customPath, "/"), nil
	}
	if strings.HasSuffix(base, "/chat/completions") {
		return base, nil
	}
	if strings.HasSuffix(base, "/v1") {
		return base + "/chat/completions", nil
	}
	return base + "/v1/chat/completions", nil
}

func doJSON(ctx context.Context, req *http.Request, output interface{}) error {
	response, err := (&http.Client{Timeout: 60 * time.Second}).Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024))
	if err != nil {
		return err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("文本渠道调用失败: %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	if err := json.Unmarshal(body, output); err != nil {
		return fmt.Errorf("解析文本渠道响应失败: %w", err)
	}
	return nil
}

func stringValue(config map[string]interface{}, key string) string {
	value, _ := config[key].(string)
	return strings.TrimSpace(value)
}
