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
	resolution, _ := params["resolution"].(string)
	sizes := map[string]map[string]string{
		"1K": {"1:1":"1024x1024","4:3":"1024x768","3:4":"768x1024","3:2":"1152x768","2:3":"768x1152","5:4":"960x768","4:5":"768x960","16:9":"1280x720","9:16":"720x1280","2:1":"1280x640","1:2":"640x1280","21:9":"1344x576","9:21":"576x1344","3:1":"1536x512","1:3":"512x1536"},
		"2K": {"1:1":"2048x2048","4:3":"2048x1536","3:4":"1536x2048","3:2":"2304x1536","2:3":"1536x2304","5:4":"2000x1600","4:5":"1600x2000","16:9":"2560x1440","9:16":"1440x2560","2:1":"2048x1024","1:2":"1024x2048","21:9":"2688x1152","9:21":"1152x2688","3:1":"3072x1024","1:3":"1024x3072"},
		"4K": {"1:1":"2880x2880","4:3":"3200x2400","3:4":"2400x3200","3:2":"3456x2304","2:3":"2304x3456","16:9":"3840x2160","9:16":"2160x3840"},
	}
	if resolution == "" { resolution = "1K" }
	if _, ok := sizes[resolution]; !ok { resolution = "1K" }
	if size := sizes[resolution][ratio]; size != "" { return size }
	return sizes[resolution]["1:1"]
}
