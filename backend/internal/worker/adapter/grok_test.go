package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/javapub/agi-platform-backend/internal/model"
)

func TestGrokAdapterSubmitsAndPollsVideoTask(t *testing.T) {
	var createPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/video/generations":
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected create method: %s", r.Method)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
				t.Fatalf("unexpected authorization: %q", got)
			}
			if err := json.NewDecoder(r.Body).Decode(&createPayload); err != nil {
				t.Fatal(err)
			}
			_, _ = w.Write([]byte(`{"data":{"id":"grok-job-1","status":"processing"}}`))
		case "/v1/video/generations/grok-job-1":
			_, _ = w.Write([]byte(`{"data":{"id":"grok-job-1","status":"completed","result":{"video_url":"https://upstream.example/video.mp4","thumbnail_url":"https://upstream.example/poster.jpg"}}}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	adapter := NewGrokAdapter(map[string]interface{}{"api_url": server.URL, "api_key": "test-key"})
	created, err := adapter.Submit(context.Background(), &GenerateRequest{
		ModelName: "grok-video-1.5",
		Type:      "video",
		Prompt:    "a cinematic city",
		Params: map[string]interface{}{
			"ratio":           "16:9",
			"duration":        "5",
			"first_frame_url": "https://assets.example/first.jpg",
			"last_frame_url":  "https://assets.example/last.jpg",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ProviderTaskID != "grok-job-1" || created.Status != "processing" {
		t.Fatalf("unexpected submitted task: %#v", created)
	}
	if createPayload["aspect_ratio"] != "16:9" || createPayload["seconds"] != "5" {
		t.Fatalf("unexpected create payload: %#v", createPayload)
	}
	refs, ok := createPayload["images"].([]interface{})
	if !ok || len(refs) != 2 {
		t.Fatalf("expected two reference URLs, got %#v", createPayload["images"])
	}

	completed, err := adapter.Poll(context.Background(), created.ProviderTaskID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != "succeeded" || completed.Result == nil || completed.Result.VideoURL != "https://upstream.example/video.mp4" {
		t.Fatalf("unexpected completed task: %#v", completed)
	}
}

func TestDiscoverGrokModelsClassifiesVideoAndTextModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"grok-image-video"},{"id":"grok-video-1.5fast"},{"id":"grok-text"}]}`))
	}))
	defer server.Close()

	models, err := discoverGrokModels(context.Background(), &model.AIProviderAccount{APIURL: server.URL, APIKey: "test-key"})
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 3 || models[0].Name != "grok-image-video" || models[0].Type != "video" || models[1].Name != "grok-video-1.5fast" || models[1].Type != "video" || models[2].Name != "grok-text" || models[2].Type != "text" {
		t.Fatalf("unexpected discovered models: %#v", models)
	}
}
