package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGeminiAdapterGeneratesImageWithConfiguredSize(t *testing.T) {
	var payload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta/models/gemini-3-pro-image-preview:generateContent" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("key") != "test-key" {
			t.Fatalf("unexpected API key query")
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"inlineData":{"mimeType":"image/png","data":"aW1hZ2U="}}]}}]}`))
	}))
	defer server.Close()

	adapter := NewGeminiAdapter(map[string]interface{}{"api_url": server.URL, "api_key": "test-key"})
	result, err := adapter.Generate(context.Background(), &GenerateRequest{
		ModelName: "gemini-3-pro-image-preview", Type: "image", Prompt: "a studio portrait",
		Params: map[string]interface{}{"ratio": "16:9", "resolution": "2K"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ImageBase64 != "aW1hZ2U=" {
		t.Fatalf("unexpected result: %#v", result)
	}
	config, ok := payload["generationConfig"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing generation config: %#v", payload)
	}
	imageConfig := config["imageConfig"].(map[string]interface{})
	if imageConfig["aspectRatio"] != "16:9" || imageConfig["imageSize"] != "2K" {
		t.Fatalf("unexpected image config: %#v", imageConfig)
	}
}
