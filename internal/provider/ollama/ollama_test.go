package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/therxl/ru-llm-router/internal/provider"
)

func TestProviderChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var request chatRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
		}

		if request.Model != "qwen2.5:3b" {
			t.Errorf("unexpected model: %s", request.Model)
		}
		if request.Stream {
			t.Error("expected stream to be false")
		}
		if len(request.Messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(request.Messages))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"qwen2.5:3b","message":{"role":"assistant","content":"Привет! Я Qwen."},"done":true}`))
	}))
	defer server.Close()

	p := New()
	result, err := p.Chat(
		context.Background(),
		provider.Connection{
			Model:   "qwen2.5:3b",
			BaseURL: server.URL,
		},
		provider.ChatRequest{
			Messages: []provider.Message{
				{Role: "user", Content: "Привет!"},
			},
		},
	)

	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if result.Model != "qwen2.5:3b" {
		t.Errorf("unexpected response model: %s", result.Model)
	}
	if result.Content != "Привет! Я Qwen." {
		t.Errorf("unexpected response content: %s", result.Content)
	}
}

func TestProviderChatHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model not found", http.StatusNotFound)
	}))
	defer server.Close()

	p := New()
	_, err := p.Chat(
		context.Background(),
		provider.Connection{
			Model:   "missing-model",
			BaseURL: server.URL,
		},
		provider.ChatRequest{},
	)

	if err == nil {
		t.Fatal("expected an error for HTTP 404")
	}
}

func TestChatURLValidation(t *testing.T) {
	_, err := chatURL("")
	if err == nil {
		t.Fatal("expected an error for empty URL")
	}

	_, err = chatURL("not-a-valid-url")
	if err == nil {
		t.Fatal("expected an error for invalid URL")
	}
}
