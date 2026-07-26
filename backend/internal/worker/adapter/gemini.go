package adapter

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GeminiAdapter implements Gemini's generateContent image protocol. Image
// inputs are fetched from our object storage and sent as inline data because
// Gemini does not accept an arbitrary public image URL in a content part.
type GeminiAdapter struct {
	config map[string]interface{}
	client *http.Client
}

func NewGeminiAdapter(config map[string]interface{}) *GeminiAdapter {
	return &GeminiAdapter{config: config, client: &http.Client{Timeout: 5 * time.Minute}}
}

func (a *GeminiAdapter) Generate(ctx context.Context, input *GenerateRequest) (*GenerateResponse, error) {
	if input.Type != "image" {
		return nil, fmt.Errorf("Gemini 渠道暂不支持 %s 生成", input.Type)
	}
	endpoint, err := a.endpoint(input.ModelName)
	if err != nil {
		return nil, err
	}
	parts := []map[string]interface{}{{"text": input.Prompt}}
	if referenceURL := configString(input.Params, "reference_image_url", ""); referenceURL != "" {
		part, err := a.referencePart(ctx, referenceURL)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	generationConfig := map[string]interface{}{
		"responseModalities": []string{"IMAGE"},
		"imageConfig": map[string]interface{}{
			"aspectRatio": configString(input.Params, "ratio", "1:1"),
		},
	}
	if size := configString(input.Params, "resolution", ""); size != "" {
		generationConfig["imageConfig"].(map[string]interface{})["imageSize"] = size
	}
	payload, err := json.Marshal(map[string]interface{}{
		"contents":         []map[string]interface{}{{"role": "user", "parts": parts}},
		"generationConfig": generationConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("编码 Gemini 图片请求失败: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 15*1024*1024))
	if err != nil {
		return nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Gemini 图片生成失败: %s: %s", response.Status, strings.TrimSpace(string(data)))
	}
	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					InlineData struct {
						Data string `json:"data"`
					} `json:"inlineData"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("解析 Gemini 图片响应失败: %w", err)
	}
	for _, candidate := range result.Candidates {
		for _, part := range candidate.Content.Parts {
			if strings.TrimSpace(part.InlineData.Data) != "" {
				return &GenerateResponse{ImageBase64: part.InlineData.Data}, nil
			}
		}
	}
	return nil, fmt.Errorf("Gemini 未返回图片结果")
}

func (a *GeminiAdapter) endpoint(modelName string) (string, error) {
	base := strings.TrimRight(configString(a.config, "api_url", ""), "/")
	if base == "" {
		return "", fmt.Errorf("渠道未配置 API 地址")
	}
	if !strings.Contains(base, "/v1beta") {
		base += "/v1beta"
	}
	endpoint := base + "/models/" + url.PathEscape(modelName) + ":generateContent"
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set("key", configString(a.config, "api_key", ""))
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func (a *GeminiAdapter) referencePart(ctx context.Context, referenceURL string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, referenceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建参考图下载请求失败: %w", err)
	}
	response, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("下载参考图失败: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("下载参考图失败: %s", response.Status)
	}
	if !strings.HasPrefix(response.Header.Get("Content-Type"), "image/") {
		return nil, fmt.Errorf("参考图不是有效图片")
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 10*1024*1024+1))
	if err != nil {
		return nil, fmt.Errorf("读取参考图失败: %w", err)
	}
	if len(data) > 10*1024*1024 {
		return nil, fmt.Errorf("参考图不能超过 10MB")
	}
	mimeType := strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0])
	return map[string]interface{}{"inlineData": map[string]string{
		"mimeType": mimeType, "data": base64.StdEncoding.EncodeToString(data),
	}}, nil
}
