package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/therxl/ru-llm-router/internal/provider"
	"github.com/therxl/ru-llm-router/internal/provider/mock"
)

type testRouter struct{ p provider.Provider }

func (t testRouter) Chat(ctx context.Context, req provider.ChatRequest) (provider.ChatResponse, error) {
	return t.p.Chat(ctx, req.Model, req)
}

func TestChatCompletions(t *testing.T) {
	handler := NewHandler(testRouter{p: mock.New()})
	body := []byte(`{"model":"mock-model","messages":[{"role":"user","content":"hello"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["object"] != "chat.completion" {
		t.Fatalf("unexpected object: %#v", result["object"])
	}
}

func TestChatCompletionsRejectsInvalidRequest(t *testing.T) {
	handler := NewHandler(testRouter{p: mock.New()})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"model":""}`))
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
