package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/therxl/ru-llm-router/internal/provider"
)

func TestChatSendsOpenAICompatibleRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-secret" {
			t.Errorf("unexpected authorization header: %q", got)
		}
		var body struct {
			Model    string             `json:"model"`
			Messages []provider.Message `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if body.Model != "test-model" || len(body.Messages) != 1 || body.Messages[0].Content != "hello" {
			t.Errorf("unexpected request body: %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"test-model","choices":[{"message":{"content":"hello back"}}]}`))
	}))
	defer server.Close()

	got, err := New().Chat(context.Background(), provider.Connection{
		Model: "test-model", BaseURL: server.URL + "/v1", APIKey: "test-secret",
	}, provider.ChatRequest{Model: "test-model", Messages: []provider.Message{{Role: "user", Content: "hello"}}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Model != "test-model" || got.Content != "hello back" {
		t.Fatalf("unexpected response: %#v", got)
	}
}

func TestChatRejectsInvalidBaseURL(t *testing.T) {
	_, err := New().Chat(context.Background(), provider.Connection{Model: "model", BaseURL: "not-a-url"}, provider.ChatRequest{})
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url error, got %v", err)
	}
}
