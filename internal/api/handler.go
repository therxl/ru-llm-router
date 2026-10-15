package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/therxl/ru-llm-router/internal/provider"
)

type ChatRouter interface {
	Chat(ctx context.Context, req provider.ChatRequest) (provider.ChatResponse, error)
}

type Handler struct {
	router ChatRouter
}

func NewHandler(r ChatRouter) *Handler { return &Handler{router: r} }

type chatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
}

type choice struct {
	Index        int             `json:"index"`
	Message      responseMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

type responseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type errorResponse struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is supported")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var req provider.ChatRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "request body must be valid JSON")
		return
	}
	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}
	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "messages must not be empty")
		return
	}
	for _, msg := range req.Messages {
		if strings.TrimSpace(msg.Role) == "" || msg.Content == "" {
			writeError(w, http.StatusBadRequest, "invalid_request_error", "each message requires role and content")
			return
		}
	}

	resp, err := h.router.Chat(r.Context(), req)
	if err != nil {
		if strings.HasPrefix(err.Error(), "unknown model or connection") {
			writeError(w, http.StatusNotFound, "model_not_found", err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, "provider_error", "provider request failed")
		return
	}

	now := time.Now()
	requestID := fmt.Sprintf("req_%d", now.UnixNano())
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	_ = json.NewEncoder(w).Encode(chatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl_%d", now.UnixNano()),
		Object:  "chat.completion",
		Created: now.Unix(),
		Model:   resp.Model,
		Choices: []choice{{Index: 0, Message: responseMessage{Role: "assistant", Content: resp.Content}, FinishReason: "stop"}},
	})
}

func writeError(w http.ResponseWriter, status int, typ, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: errorDetail{Message: message, Type: typ}})
}
