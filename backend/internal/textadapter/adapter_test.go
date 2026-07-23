package textadapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAICompatibleComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("missing authorization header")
		}
		var body struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Model != "gpt-4.1-mini" || len(body.Messages) != 2 || body.Messages[0].Content != "system" || body.Messages[1].Content != "user" {
			t.Fatalf("unexpected request body: %#v", body)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"choices":[{"message":{"content":"  optimized prompt  "}}]}`))
	}))
	defer server.Close()

	adapter := &openAICompatible{config: map[string]interface{}{"api_url": server.URL, "api_key": "test-key"}}
	result, err := adapter.Complete(context.Background(), &Request{ModelName: "gpt-4.1-mini", SystemPrompt: "system", Prompt: "user"})
	if err != nil {
		t.Fatal(err)
	}
	if result != "optimized prompt" {
		t.Fatalf("unexpected result: %q", result)
	}
}
