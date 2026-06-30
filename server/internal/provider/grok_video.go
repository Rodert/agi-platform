package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type GrokVideoProvider struct{}

func NewGrokVideoProvider() *GrokVideoProvider {
	return &GrokVideoProvider{}
}

func (p *GrokVideoProvider) Type() string {
	return "grok-video"
}

func (p *GrokVideoProvider) SupportsImageGeneration() bool {
	return false
}

func (p *GrokVideoProvider) Generate(_ context.Context, _ ImageRequest) (*ImageResult, error) {
	return nil, errors.New("grok-video provider does not support image generation")
}

func (p *GrokVideoProvider) CreateVideo(ctx context.Context, req VideoRequest) (*VideoCreateResult, error) {
	if strings.TrimSpace(req.BaseURL) == "" {
		return nil, errors.New("provider base_url is required")
	}
	if strings.TrimSpace(req.APIKey) == "" {
		return nil, errors.New("provider api key is required")
	}
	if len(req.Videos) > 0 || len(req.Audios) > 0 {
		return nil, fmt.Errorf("grok-video provider only supports reference images")
	}
	if err := validateGrokVideoRequest(req); err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"model":        req.Model,
		"prompt":       req.Prompt,
		"seconds":      adjustedGrokSeconds(req),
		"aspect_ratio": req.AspectRatio,
		"resolution":   stringExtra(req.Extra, "resolution", "720p"),
	}
	if len(req.Images) > 0 {
		payload[stringExtra(req.Extra, "image_field", "image_urls")] = req.Images
	}
	for key, value := range payloadExtra(req.Extra) {
		payload[key] = value
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	raw, err := doProviderJSON(ctx, http.MethodPost, grokVideoGenerationsEndpoint(req.BaseURL), req.APIKey, "application/json", body, req.TimeoutSeconds)
	if err != nil {
		return nil, err
	}
	var parsed grokVideoStatusEnvelope
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decode grok video response: %w", err)
	}
	taskID := firstNonEmpty(parsed.TaskID, parsed.ID, parsed.Data.TaskID)
	if taskID == "" {
		return nil, errors.New("provider response did not include video task id")
	}
	return &VideoCreateResult{
		TaskID:      taskID,
		Status:      normalizeGrokVideoStatus(firstNonEmpty(parsed.Status, parsed.Data.Status)),
		URL:         grokVideoResponseURL(parsed),
		RawResponse: compactRawResponse(raw),
	}, nil
}

func (p *GrokVideoProvider) GetVideo(ctx context.Context, req VideoStatusRequest) (*VideoStatusResult, error) {
	raw, err := doProviderJSON(ctx, http.MethodGet, grokVideoTaskEndpoint(req.BaseURL, req.TaskID), req.APIKey, "application/json", nil, req.TimeoutSeconds)
	if err != nil {
		return nil, err
	}
	var parsed grokVideoStatusEnvelope
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decode grok video status: %w", err)
	}
	if parsed.Code != "" && !strings.EqualFold(parsed.Code, "success") {
		return nil, NewResponseError(fmt.Sprintf("provider request failed: code=%s message=%s", parsed.Code, parsed.Message), compactRawResponse(raw))
	}
	result := &VideoStatusResult{
		TaskID:       firstNonEmpty(parsed.Data.TaskID, parsed.TaskID, parsed.ID, req.TaskID),
		Status:       normalizeGrokVideoStatus(firstNonEmpty(parsed.Data.Status, parsed.Data.Data.Status, parsed.Status)),
		Progress:     grokVideoProgress(parsed),
		URL:          grokVideoResponseURL(parsed),
		ErrorMessage: parsed.Data.FailReason,
		RawResponse:  compactRawResponse(raw),
	}
	if result.Status == "failed" && result.ErrorMessage == "" {
		result.ErrorMessage = parsed.Message
	}
	return result, nil
}

func (p *GrokVideoProvider) DownloadVideo(ctx context.Context, req VideoContentRequest) ([]byte, string, error) {
	raw, err := doProviderJSON(ctx, http.MethodGet, grokVideoTaskEndpoint(req.BaseURL, req.TaskID), req.APIKey, "application/json", nil, req.TimeoutSeconds)
	if err != nil {
		return nil, "", err
	}
	var parsed grokVideoStatusEnvelope
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, "", fmt.Errorf("decode grok video status before download: %w", err)
	}
	resultURL := grokVideoResponseURL(parsed)
	if resultURL == "" {
		return nil, "", NewResponseError("provider response did not include result_url", compactRawResponse(raw))
	}
	return downloadProviderContent(ctx, resultURL, req.TimeoutSeconds)
}

type grokVideoStatusEnvelope struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Code      string   `json:"code"`
	Status    string   `json:"status"`
	Message   string   `json:"message"`
	TaskID    string   `json:"task_id"`
	Progress  int      `json:"progress"`
	VideoURL  string   `json:"video_url"`
	ResultURL string   `json:"result_url"`
	Output    []string `json:"output"`
	Video     struct {
		URL string `json:"url"`
	} `json:"video"`
	Data struct {
		TaskID     string `json:"task_id"`
		Status     string `json:"status"`
		Progress   string `json:"progress"`
		ResultURL  string `json:"result_url"`
		VideoURL   string `json:"video_url"`
		URL        string `json:"url"`
		FailReason string `json:"fail_reason"`
		Video      struct {
			URL string `json:"url"`
		} `json:"video"`
		Output []string `json:"output"`
		Data   struct {
			Status   string `json:"status"`
			Progress int    `json:"progress"`
		} `json:"data"`
	} `json:"data"`
}

func validateGrokVideoRequest(req VideoRequest) error {
	imageCount := len(req.Images)
	maxImages := intExtra(req.Extra, "max_reference_images", 7)
	if imageCount > maxImages {
		return fmt.Errorf("grok-video provider supports at most %d reference images", maxImages)
	}
	if boolExtra(req.Extra, "require_exactly_one_image", false) && imageCount != 1 {
		return errors.New("grok-video provider requires exactly one reference image")
	}
	return nil
}

func adjustedGrokSeconds(req VideoRequest) int {
	seconds := req.Seconds
	if seconds <= 0 {
		seconds = 4
	}
	if len(req.Images) > 1 {
		maxSeconds := intExtra(req.Extra, "multi_image_max_seconds", 10)
		if seconds > maxSeconds {
			return maxSeconds
		}
	}
	return seconds
}

func normalizeGrokVideoStatus(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "SUCCESS", "SUCCEEDED", "COMPLETED", "COMPLETE", "DONE":
		return "succeeded"
	case "FAILURE", "FAILED", "FAIL":
		return "failed"
	case "IN_PROGRESS":
		return "running"
	case "SUBMITTED", "QUEUED", "NOT_START", "":
		return "pending"
	default:
		return normalizeVideoStatus(status)
	}
}

func grokVideoResponseURL(parsed grokVideoStatusEnvelope) string {
	return firstNonEmpty(
		parsed.ResultURL,
		parsed.VideoURL,
		parsed.URL,
		parsed.Video.URL,
		parsed.Data.ResultURL,
		parsed.Data.VideoURL,
		parsed.Data.URL,
		parsed.Data.Video.URL,
		firstString(parsed.Output),
		firstString(parsed.Data.Output),
	)
}

func grokVideoProgress(parsed grokVideoStatusEnvelope) int {
	if parsed.Data.Progress != "" {
		return parseProgressPercent(parsed.Data.Progress)
	}
	if parsed.Progress > 0 {
		if parsed.Progress > 100 {
			return 100
		}
		return parsed.Progress
	}
	if parsed.Data.Data.Progress > 0 {
		if parsed.Data.Data.Progress > 100 {
			return 100
		}
		return parsed.Data.Data.Progress
	}
	return 0
}

func firstString(values []string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func parseProgressPercent(value string) int {
	trimmed := strings.TrimSuffix(strings.TrimSpace(value), "%")
	if trimmed == "" {
		return 0
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0
	}
	if parsed < 0 {
		return 0
	}
	if parsed > 100 {
		return 100
	}
	return parsed
}

func grokVideoGenerationsEndpoint(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(trimmed, "/video/generations") {
		return trimmed
	}
	if strings.HasSuffix(trimmed, "/v1") {
		return trimmed + "/video/generations"
	}
	return trimmed + "/v1/video/generations"
}

func grokVideoTaskEndpoint(baseURL string, taskID string) string {
	return grokVideoGenerationsEndpoint(baseURL) + "/" + url.PathEscape(taskID)
}

func doProviderJSON(ctx context.Context, method string, endpoint string, apiKey string, contentType string, body []byte, timeoutSeconds int) ([]byte, error) {
	timeout := time.Duration(defaultTimeoutSeconds(timeoutSeconds)) * time.Second
	httpReq, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Accept", "application/json")
	if body != nil {
		httpReq.Header.Set("Content-Type", contentType)
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, NewHTTPResponseError("provider request failed", resp.StatusCode, compactRawResponse(raw))
	}
	if err := providerBusinessError(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func downloadProviderContent(ctx context.Context, endpoint string, timeoutSeconds int) ([]byte, string, error) {
	timeout := time.Duration(defaultTimeoutSeconds(timeoutSeconds)) * time.Second
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", err
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 512<<20))
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", NewHTTPResponseError("provider content failed", resp.StatusCode, compactRawResponse(raw))
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "video/mp4"
	}
	return raw, contentType, nil
}

func boolExtra(extra map[string]interface{}, key string, fallback bool) bool {
	value, ok := extra[key]
	if !ok {
		return fallback
	}
	parsed, ok := value.(bool)
	if !ok {
		return fallback
	}
	return parsed
}

func intExtra(extra map[string]interface{}, key string, fallback int) int {
	value, ok := extra[key]
	if !ok {
		return fallback
	}
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil {
			return parsed
		}
	}
	return fallback
}
