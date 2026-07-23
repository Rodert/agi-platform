package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIAdapter implements the OpenAI Images API and compatible gateways.
// Gateways only need to expose POST /v1/images/generations with Bearer auth.
type OpenAIAdapter struct { config map[string]interface{} }

func NewOpenAIAdapter(config map[string]interface{}) *OpenAIAdapter { return &OpenAIAdapter{config: config} }

func (a *OpenAIAdapter) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	if req.Type != "image" { return nil, fmt.Errorf("OpenAI 兼容渠道暂不支持 %s 生成", req.Type) }
	endpoint, err := a.endpoint()
	if err != nil { return nil, err }
	body := map[string]interface{}{"model": req.ModelName, "prompt": req.Prompt, "n": 1, "response_format": "url"}
	if size := openAIImageSize(req.Params); size != "" { body["size"] = size }
	if quality, ok := req.Params["quality"].(string); ok && quality != "" { body["quality"] = quality }
	payload, err := json.Marshal(body)
	if err != nil { return nil, err }
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil { return nil, err }
	httpReq.Header.Set("Content-Type", "application/json")
	if key, _ := a.config["api_key"].(string); key != "" { httpReq.Header.Set("Authorization", "Bearer "+key) }
	client := &http.Client{Timeout: 5 * time.Minute}
	response, err := client.Do(httpReq)
	if err != nil { return nil, err }
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 10*1024*1024))
	if err != nil { return nil, err }
	if response.StatusCode < 200 || response.StatusCode >= 300 { return nil, fmt.Errorf("上游图片生成失败: %s: %s", response.Status, strings.TrimSpace(string(data))) }
	var result struct { Data []struct { URL string `json:"url"`; B64JSON string `json:"b64_json"` } `json:"data"` }
	if err := json.Unmarshal(data, &result); err != nil { return nil, fmt.Errorf("解析上游图片响应失败: %w", err) }
	if len(result.Data) == 0 { return nil, fmt.Errorf("上游未返回图片结果") }
	return &GenerateResponse{ImageURL: result.Data[0].URL, ImageBase64: result.Data[0].B64JSON, ThumbnailURL: result.Data[0].URL}, nil
}

func (a *OpenAIAdapter) endpoint() (string, error) {
	raw, _ := a.config["api_url"].(string)
	raw = strings.TrimRight(raw, "/")
	if raw == "" { return "", fmt.Errorf("渠道未配置 API 地址") }
	if strings.Contains(raw, "/images/") { return raw, nil }
	return raw + "/v1/images/generations", nil
}

func openAIImageSize(params map[string]interface{}) string {
	ratio, _ := params["ratio"].(string)
	switch ratio { case "16:9", "3:2": return "1536x1024"; case "9:16", "2:3": return "1024x1536"; default: return "1024x1024" }
}
