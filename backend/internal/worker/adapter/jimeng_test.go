package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/javapub/agi-platform-backend/internal/model"
)

func TestJimengInternationalAdapterExplainsHTMLResponse(t *testing.T) {
	adapter := NewJimengInternationalAdapter(nil)
	_, err := adapter.parseTask([]byte("<!doctype html><html></html>"), "")
	if err == nil || !strings.Contains(err.Error(), "https://zz1cc.cc.cd/v1") {
		t.Fatalf("unexpected HTML response error: %v", err)
	}
}

func TestJimengInternationalAdapterSubmitsAndPollsVideoTask(t *testing.T) {
	var createPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization: %q", got)
		}
		switch request.URL.Path {
		case "/v1/videos":
			if request.Method != http.MethodPost {
				t.Fatalf("unexpected create method: %s", request.Method)
			}
			if err := json.NewDecoder(request.Body).Decode(&createPayload); err != nil {
				t.Fatal(err)
			}
			_, _ = w.Write([]byte(`{"id":"jimeng-job-1","status":"queued"}`))
		case "/v1/videos/jimeng-job-1":
			_, _ = w.Write([]byte(`{"id":"jimeng-job-1","status":"completed","progress":100}`))
		default:
			t.Fatalf("unexpected request path: %s", request.URL.Path)
		}
	}))
	defer server.Close()

	adapter := NewJimengInternationalAdapter(map[string]interface{}{"api_url": server.URL + "/v1", "api_key": "test-key"})
	created, err := adapter.Submit(context.Background(), &GenerateRequest{
		ModelName: "video-ds-2.0-fast", Type: "video", Prompt: "a cinematic city",
		Params: map[string]interface{}{"ratio": "9:16", "duration": "15", "reference_image_urls": []string{"https://assets.example/reference.jpg"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ProviderTaskID != "jimeng-job-1" || created.Status != "queued" {
		t.Fatalf("unexpected submitted task: %#v", created)
	}
	if createPayload["aspect_ratio"] != "9:16" || createPayload["seconds"] != "15" {
		t.Fatalf("unexpected create payload: %#v", createPayload)
	}
	if images, ok := createPayload["images"].([]interface{}); !ok || len(images) != 1 {
		t.Fatalf("unexpected reference images: %#v", createPayload["images"])
	}

	completed, err := adapter.Poll(context.Background(), created.ProviderTaskID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != "succeeded" || completed.Result == nil || completed.Result.VideoURL != server.URL+"/v1/videos/jimeng-job-1/content" || completed.Result.VideoHeaders["Authorization"] != "Bearer test-key" {
		t.Fatalf("unexpected completed task: %#v", completed)
	}
}

func TestDiscoverJimengInternationalModelsClassifiesVideoModels(t *testing.T) {
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requestedPath = request.URL.Path
		if request.URL.Path != "/v1/models" {
			t.Fatalf("unexpected request path: %s", request.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"video-ds-2.0"},{"id":"video-ds-2.0-fast"},{"id":"as-sd2.0-fast"}]}`))
	}))
	defer server.Close()

	models, err := discoverJimengInternationalModels(context.Background(), &model.AIProviderAccount{APIURL: server.URL + "/v1", APIKey: "test-key", ExtraConfig: []byte(`{"models_path":"/v1/models"}`)})
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 3 {
		t.Fatalf("unexpected models: %#v", models)
	}
	if requestedPath != "/v1/models" {
		t.Fatalf("unexpected models endpoint: %s", requestedPath)
	}
	for _, item := range models {
		if item.Type != "video" {
			t.Fatalf("expected video model, got %#v", item)
		}
	}
}
