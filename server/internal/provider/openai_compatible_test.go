package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAICompatibleGenerateRejectsBusinessErrorCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":400,"message":"bad prompt","data":null}`))
	}))
	defer server.Close()

	provider := NewOpenAICompatibleProvider("openai-compatible")
	_, err := provider.Generate(context.Background(), ImageRequest{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "gpt-image-2",
		Prompt:  "test",
		Size:    "1024x1024",
	})
	if err == nil || !strings.Contains(err.Error(), "code=400") || !strings.Contains(err.Error(), "bad prompt") {
		t.Fatalf("expected business error, got %v", err)
	}
}

func TestOpenAICompatibleCreateVideoRejectsBusinessErrorCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":400,"message":"invalid video request","data":null}`))
	}))
	defer server.Close()

	provider := NewOpenAICompatibleProvider("openai-compatible")
	_, err := provider.CreateVideo(context.Background(), VideoRequest{
		BaseURL:     server.URL,
		APIKey:      "test-key",
		Model:       "video-ds-2.0-fast",
		Prompt:      "test",
		Seconds:     5,
		AspectRatio: "9:16",
	})
	if err == nil || !strings.Contains(err.Error(), "code=400") || !strings.Contains(err.Error(), "invalid video request") {
		t.Fatalf("expected business error, got %v", err)
	}
}

func TestOpenAICompatibleGetVideoRejectsBusinessErrorCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":400,"message":"fail_to_fetch_task","data":null}`))
	}))
	defer server.Close()

	provider := NewOpenAICompatibleProvider("openai-compatible")
	_, err := provider.GetVideo(context.Background(), VideoStatusRequest{
		BaseURL: server.URL,
		APIKey:  "test-key",
		TaskID:  "task_1",
	})
	if err == nil || !strings.Contains(err.Error(), "code=400") || !strings.Contains(err.Error(), "fail_to_fetch_task") {
		t.Fatalf("expected business error, got %v", err)
	}
}
