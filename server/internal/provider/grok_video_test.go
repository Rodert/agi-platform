package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGrokVideoProviderCreateAndPoll(t *testing.T) {
	var createPath string
	var createPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/video/generations":
			createPath = r.URL.Path
			if err := json.NewDecoder(r.Body).Decode(&createPayload); err != nil {
				t.Fatalf("decode create payload: %v", err)
			}
			_, _ = w.Write([]byte(`{"id":"task_1","task_id":"task_1","status":"queued","progress":0}`))
		case "/v1/video/generations/task_1":
			_, _ = w.Write([]byte(`{"code":"success","message":"","data":{"task_id":"task_1","status":"SUCCESS","progress":"100%","result_url":"https://example.com/out.mp4","fail_reason":""}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := NewGrokVideoProvider()
	created, err := provider.CreateVideo(context.Background(), VideoRequest{
		BaseURL:     server.URL,
		APIKey:      "test-key",
		Model:       "grok-image-video",
		Prompt:      "test",
		Seconds:     12,
		AspectRatio: "16:9",
		Images:      []string{"https://example.com/1.png", "https://example.com/2.png"},
		Extra: map[string]interface{}{
			"resolution":              "720p",
			"multi_image_max_seconds": 10,
		},
	})
	if err != nil {
		t.Fatalf("create video: %v", err)
	}
	if createPath != "/v1/video/generations" {
		t.Fatalf("unexpected create path %s", createPath)
	}
	if created.TaskID != "task_1" || created.Status != "pending" {
		t.Fatalf("unexpected create result %+v", created)
	}
	if createPayload["seconds"] != float64(10) {
		t.Fatalf("expected multi-image seconds capped to 10, got %#v", createPayload["seconds"])
	}
	if _, ok := createPayload["image_urls"]; !ok {
		t.Fatalf("expected image_urls in payload: %#v", createPayload)
	}

	status, err := provider.GetVideo(context.Background(), VideoStatusRequest{
		BaseURL: server.URL,
		APIKey:  "test-key",
		TaskID:  "task_1",
	})
	if err != nil {
		t.Fatalf("get video: %v", err)
	}
	if status.Status != "succeeded" || status.URL == "" || status.Progress != 100 {
		t.Fatalf("unexpected status %+v", status)
	}
}

func TestGrokVideoProviderPollFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"success","message":"","data":{"task_id":"task_1","status":"FAILURE","progress":"100%","result_url":"","fail_reason":"image fetch failed"}}`))
	}))
	defer server.Close()

	provider := NewGrokVideoProvider()
	status, err := provider.GetVideo(context.Background(), VideoStatusRequest{
		BaseURL: server.URL,
		APIKey:  "test-key",
		TaskID:  "task_1",
	})
	if err != nil {
		t.Fatalf("get video: %v", err)
	}
	if status.Status != "failed" || !strings.Contains(status.ErrorMessage, "image fetch failed") {
		t.Fatalf("unexpected failure status %+v", status)
	}
}

func TestGrokVideoProviderParsesCompletedURLVariants(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"task_WTmyjQzET5tlFIyep4SrrQbPWBpThyKr",
			"url":"https://example.com/top.mp4",
			"code":"success",
			"data":{
				"task_id":"task_WTmyjQzET5tlFIyep4SrrQbPWBpThyKr",
				"status":"SUCCESS",
				"progress":"100%",
				"result_url":"https://example.com/data-result.mp4"
			},
			"video":{"url":"https://example.com/video.mp4"},
			"output":["https://example.com/output.mp4"],
			"status":"completed",
			"task_id":"task_WTmyjQzET5tlFIyep4SrrQbPWBpThyKr",
			"progress":100,
			"video_url":"https://example.com/video-url.mp4",
			"result_url":"https://example.com/result.mp4"
		}`))
	}))
	defer server.Close()

	provider := NewGrokVideoProvider()
	created, err := provider.CreateVideo(context.Background(), VideoRequest{
		BaseURL:     server.URL,
		APIKey:      "test-key",
		Model:       "grok-image-video",
		Prompt:      "test",
		Seconds:     15,
		AspectRatio: "16:9",
	})
	if err != nil {
		t.Fatalf("create video: %v", err)
	}
	if created.Status != "succeeded" {
		t.Fatalf("expected succeeded, got %s", created.Status)
	}
	if created.URL != "https://example.com/result.mp4" {
		t.Fatalf("unexpected create url %s", created.URL)
	}

	status, err := provider.GetVideo(context.Background(), VideoStatusRequest{
		BaseURL: server.URL,
		APIKey:  "test-key",
		TaskID:  created.TaskID,
	})
	if err != nil {
		t.Fatalf("get video: %v", err)
	}
	if status.Status != "succeeded" || status.Progress != 100 || status.URL != "https://example.com/result.mp4" {
		t.Fatalf("unexpected status %+v", status)
	}
}
