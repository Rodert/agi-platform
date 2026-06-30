package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OpenAICompatibleProvider struct {
	providerType string
}

func NewOpenAICompatibleProvider(providerType string) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{providerType: providerType}
}

func (p *OpenAICompatibleProvider) Type() string {
	return p.providerType
}

func (p *OpenAICompatibleProvider) Generate(ctx context.Context, req ImageRequest) (*ImageResult, error) {
	if strings.TrimSpace(req.BaseURL) == "" {
		return nil, errors.New("provider base_url is required")
	}
	if strings.TrimSpace(req.APIKey) == "" {
		return nil, errors.New("provider api key is required")
	}

	endpoint := imageGenerationsEndpoint(req.BaseURL)
	body, contentType, err := buildGenerationBody(req)
	if err != nil {
		return nil, err
	}
	if hasMultipartReference(req.ReferenceImages) {
		endpoint = imageEditsEndpoint(req.BaseURL)
	}

	timeout := time.Duration(defaultTimeoutSeconds(req.TimeoutSeconds)) * time.Second
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, NewHTTPResponseError("provider request failed", resp.StatusCode, compactRawResponse(raw))
	}
	if err := providerBusinessError(raw); err != nil {
		return nil, err
	}

	var parsed openAIImageResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decode provider response: %w", err)
	}
	images := make([]ImageItem, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		url := item.URL
		if url == "" && item.B64JSON != "" {
			url = "data:image/png;base64," + item.B64JSON
		}
		if url == "" {
			continue
		}
		images = append(images, ImageItem{URL: url})
	}
	if len(images) == 0 {
		return nil, errors.New("provider response did not include image urls")
	}

	return &ImageResult{
		ProviderTaskID: firstNonEmpty(parsed.ID, parsed.TaskID),
		Status:         firstNonEmpty(parsed.Status, "succeeded"),
		Images:         images,
		RawResponse:    compactImageResponseRaw(parsed, raw),
	}, nil
}

func (p *OpenAICompatibleProvider) CreateVideo(ctx context.Context, req VideoRequest) (*VideoCreateResult, error) {
	if strings.TrimSpace(req.BaseURL) == "" {
		return nil, errors.New("provider base_url is required")
	}
	if strings.TrimSpace(req.APIKey) == "" {
		return nil, errors.New("provider api key is required")
	}
	payload := map[string]interface{}{
		"model":        req.Model,
		"prompt":       req.Prompt,
		"seconds":      req.Seconds,
		"aspect_ratio": req.AspectRatio,
	}
	if len(req.Images) > 0 {
		payload["images"] = req.Images
	}
	if len(req.Videos) > 0 {
		payload["videos"] = req.Videos
	}
	if len(req.Audios) > 0 {
		payload["audios"] = req.Audios
	}
	for key, value := range payloadExtra(req.Extra) {
		payload[key] = value
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	raw, err := p.doJSON(ctx, http.MethodPost, videoEndpoint(req.BaseURL), req.APIKey, "application/json", body, req.TimeoutSeconds)
	if err != nil {
		return nil, err
	}
	var parsed openAIVideoResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decode provider video response: %w", err)
	}
	taskID := firstNonEmpty(parsed.ID, parsed.TaskID)
	if taskID == "" {
		return nil, errors.New("provider response did not include video task id")
	}
	return &VideoCreateResult{
		TaskID:      taskID,
		Status:      firstNonEmpty(parsed.Status, "pending"),
		RawResponse: string(raw),
	}, nil
}

func (p *OpenAICompatibleProvider) GetVideo(ctx context.Context, req VideoStatusRequest) (*VideoStatusResult, error) {
	raw, err := p.doJSON(ctx, http.MethodGet, videoTaskEndpoint(req.BaseURL, req.TaskID), req.APIKey, "application/json", nil, req.TimeoutSeconds)
	if err != nil {
		return nil, err
	}
	var parsed openAIVideoResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decode provider video status: %w", err)
	}
	result := &VideoStatusResult{
		TaskID:      firstNonEmpty(parsed.ID, parsed.TaskID, req.TaskID),
		Status:      normalizeVideoStatus(parsed.Status),
		Progress:    parsed.Progress,
		URL:         firstNonEmpty(parsed.URL, parsed.OutputURL),
		RawResponse: string(raw),
	}
	if parsed.Error != nil {
		result.ErrorCode = parsed.Error.Code
		result.ErrorMessage = parsed.Error.Message
	}
	return result, nil
}

func (p *OpenAICompatibleProvider) DownloadVideo(ctx context.Context, req VideoContentRequest) ([]byte, string, error) {
	endpoint := videoTaskEndpoint(req.BaseURL, req.TaskID) + "/content"
	timeout := time.Duration(defaultTimeoutSeconds(req.TimeoutSeconds)) * time.Second
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
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
		return nil, "", NewHTTPResponseError("provider video content failed", resp.StatusCode, compactRawResponse(raw))
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "video/mp4"
	}
	return raw, contentType, nil
}

func (p *OpenAICompatibleProvider) doJSON(ctx context.Context, method string, endpoint string, apiKey string, contentType string, body []byte, timeoutSeconds int) ([]byte, error) {
	return doProviderJSON(ctx, method, endpoint, apiKey, contentType, body, timeoutSeconds)
}

type openAIImageResponse struct {
	ID     string            `json:"id"`
	TaskID string            `json:"task_id"`
	Status string            `json:"status"`
	Data   []openAIImageData `json:"data"`
}

type openAIImageData struct {
	URL     string `json:"url"`
	B64JSON string `json:"b64_json"`
}

type openAIVideoResponse struct {
	ID        string `json:"id"`
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	Progress  int    `json:"progress"`
	URL       string `json:"url"`
	OutputURL string `json:"output_url"`
	Error     *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type providerEnvelope struct {
	Code    *int            `json:"code"`
	Message string          `json:"message"`
	Error   json.RawMessage `json:"error"`
}

type providerErrorPayload struct {
	Code    string `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func providerBusinessError(raw []byte) error {
	var envelope providerEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil || envelope.Code == nil || providerBusinessCodeOK(*envelope.Code) {
		return nil
	}
	message := strings.TrimSpace(envelope.Message)
	if message == "" && len(envelope.Error) > 0 && string(envelope.Error) != "null" {
		var payload providerErrorPayload
		if err := json.Unmarshal(envelope.Error, &payload); err == nil {
			message = firstNonEmpty(strings.TrimSpace(payload.Message), strings.TrimSpace(payload.Type), strings.TrimSpace(payload.Code))
		}
		if message == "" {
			message = string(envelope.Error)
		}
	}
	if message == "" {
		message = "provider returned non-zero code"
	}
	return NewResponseError(
		fmt.Sprintf("provider request failed: code=%d message=%s", *envelope.Code, truncate(message, 1000)),
		compactRawResponse(raw),
	)
}

func providerBusinessCodeOK(code int) bool {
	return code == 0 || (code >= 200 && code < 300)
}

func compactImageResponseRaw(parsed openAIImageResponse, raw []byte) string {
	if !responseContainsB64(parsed) {
		return compactRawResponse(raw)
	}
	items := make([]map[string]interface{}, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		items = append(items, map[string]interface{}{
			"has_url":      item.URL != "",
			"has_b64_json": item.B64JSON != "",
			"b64_bytes":    len(item.B64JSON),
		})
	}
	compact := map[string]interface{}{
		"id":      parsed.ID,
		"task_id": parsed.TaskID,
		"status":  parsed.Status,
		"data":    items,
	}
	encoded, err := json.Marshal(compact)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func compactRawResponse(raw []byte) string {
	if len(raw) <= 64*1024 {
		return string(raw)
	}
	compact := map[string]interface{}{
		"truncated":    true,
		"length_bytes": len(raw),
	}
	encoded, err := json.Marshal(compact)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func responseContainsB64(parsed openAIImageResponse) bool {
	for _, item := range parsed.Data {
		if item.B64JSON != "" {
			return true
		}
	}
	return false
}

func buildGenerationBody(req ImageRequest) ([]byte, string, error) {
	if hasMultipartReference(req.ReferenceImages) {
		return buildMultipartEditBody(req)
	}

	payload := map[string]interface{}{
		"model":  req.Model,
		"prompt": req.Prompt,
		"n":      defaultImageCount(req.N),
		"size":   req.Size,
	}
	if req.NegativePrompt != "" {
		payload["negative_prompt"] = req.NegativePrompt
	}
	for key, value := range payloadExtra(req.Extra) {
		payload[key] = value
	}
	if len(req.ReferenceImages) > 0 {
		field := stringExtra(req.Extra, "reference_images_field", "image")
		values := referenceImageValues(req.ReferenceImages)
		if stringExtra(req.Extra, "reference_images_mode", "array") == "string" && len(values) == 1 {
			payload[field] = values[0]
		} else {
			payload[field] = values
		}
	}

	body, err := json.Marshal(payload)
	return body, "application/json", err
}

func buildMultipartEditBody(req ImageRequest) ([]byte, string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	fields := map[string]string{
		"model":  req.Model,
		"prompt": req.Prompt,
		"n":      fmt.Sprintf("%d", defaultImageCount(req.N)),
		"size":   req.Size,
	}
	if req.NegativePrompt != "" {
		fields["negative_prompt"] = req.NegativePrompt
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", err
		}
	}

	field := stringExtra(req.Extra, "reference_images_field", "image")
	for index, image := range req.ReferenceImages {
		dataURL := strings.TrimSpace(image.DataURL)
		if dataURL == "" {
			continue
		}
		filename := image.Filename
		if filename == "" {
			filename = fmt.Sprintf("reference-%d.png", index+1)
		}
		mimeType, data, err := decodeDataURL(dataURL)
		if err != nil {
			return nil, "", err
		}
		part, err := writer.CreateFormFile(field, filename)
		if err != nil {
			return nil, "", err
		}
		_ = mimeType
		if _, err := part.Write(data); err != nil {
			return nil, "", err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return body.Bytes(), writer.FormDataContentType(), nil
}

func imageGenerationsEndpoint(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(trimmed, "/images/generations") {
		return trimmed
	}
	if strings.HasSuffix(trimmed, "/v1") {
		return trimmed + "/images/generations"
	}
	return trimmed + "/v1/images/generations"
}

func imageEditsEndpoint(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(trimmed, "/images/edits") {
		return trimmed
	}
	if strings.HasSuffix(trimmed, "/v1") {
		return trimmed + "/images/edits"
	}
	return trimmed + "/v1/images/edits"
}

func videoEndpoint(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(trimmed, "/videos") {
		return trimmed
	}
	if strings.HasSuffix(trimmed, "/v1") {
		return trimmed + "/videos"
	}
	return trimmed + "/v1/videos"
}

func videoTaskEndpoint(baseURL string, taskID string) string {
	return videoEndpoint(baseURL) + "/" + url.PathEscape(taskID)
}

func normalizeVideoStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "succeeded", "success", "completed", "complete", "done":
		return "succeeded"
	case "failed", "fail", "error", "cancelled", "canceled":
		return "failed"
	case "running", "processing", "in_progress":
		return "running"
	default:
		return firstNonEmpty(status, "pending")
	}
}

func payloadExtra(extra map[string]interface{}) map[string]interface{} {
	value, ok := extra["payload_extra"]
	if !ok {
		return nil
	}
	extraMap, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}
	return extraMap
}

func stringExtra(extra map[string]interface{}, key string, fallback string) string {
	value, ok := extra[key].(string)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func hasMultipartReference(images []ReferenceImage) bool {
	for _, image := range images {
		if strings.TrimSpace(image.DataURL) != "" {
			return true
		}
	}
	return false
}

func referenceImageValues(images []ReferenceImage) []string {
	values := make([]string, 0, len(images))
	for _, image := range images {
		switch {
		case strings.TrimSpace(image.URL) != "":
			values = append(values, strings.TrimSpace(image.URL))
		case strings.TrimSpace(image.DataURL) != "":
			values = append(values, strings.TrimSpace(image.DataURL))
		}
	}
	return values
}

func decodeDataURL(value string) (string, []byte, error) {
	parsed, err := url.Parse(value)
	if err != nil {
		return "", nil, err
	}
	if parsed.Scheme != "data" {
		return "", nil, errors.New("reference image must be a data url")
	}
	parts := strings.SplitN(parsed.Opaque, ",", 2)
	if len(parts) != 2 {
		return "", nil, errors.New("invalid data url")
	}
	meta := parts[0]
	mimeType := "application/octet-stream"
	if semi := strings.Index(meta, ";"); semi >= 0 {
		if meta[:semi] != "" {
			mimeType = meta[:semi]
		}
	} else if meta != "" {
		mimeType = meta
	}
	if !strings.Contains(meta, ";base64") {
		return "", nil, errors.New("reference data url must be base64 encoded")
	}
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, err
	}
	return mimeType, data, nil
}

func defaultImageCount(value int) int {
	if value <= 0 {
		return 1
	}
	return value
}

func defaultTimeoutSeconds(value int) int {
	if value <= 0 {
		return 60
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}
