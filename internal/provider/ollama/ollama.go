package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/therxl/ru-llm-router/internal/provider"
)

const maxResponseBody = 4 << 20

type Provider struct {
	client *http.Client
}

func New() *Provider {
	return &Provider{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

type chatRequest struct {
	Model    string             `json:"model"`
	Messages []provider.Message `json:"messages"`
	Stream   bool               `json:"stream"`
}

type chatResponse struct {
	Model   string `json:"model"`
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

func (p *Provider) Chat(ctx context.Context, connection provider.Connection, req provider.ChatRequest) (provider.ChatResponse, error) {
	endpoint, err := chatURL(connection.BaseURL)
	if err != nil {
		return provider.ChatResponse{}, err
	}

	payload, err := json.Marshal(chatRequest{
		Model:    connection.Model,
		Messages: req.Messages,
		Stream:   false,
	})
	if err != nil {
		return provider.ChatResponse{}, fmt.Errorf("encode Ollama request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return provider.ChatResponse{}, fmt.Errorf("create Ollama request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return provider.ChatResponse{}, fmt.Errorf("send Ollama request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return provider.ChatResponse{}, fmt.Errorf("read Ollama response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return provider.ChatResponse{}, fmt.Errorf("Ollama returned HTTP %d", resp.StatusCode)
	}

	var result chatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return provider.ChatResponse{}, fmt.Errorf("decode Ollama response: %w", err)
	}

	model := result.Model
	if model == "" {
		model = connection.Model
	}

	return provider.ChatResponse{
		Model:   model,
		Content: result.Message.Content,
	}, nil
}

func chatURL(baseURL string) (string, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return "", fmt.Errorf("Ollama provider requires base_url")
	}

	parsed, err := url.Parse(baseURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", fmt.Errorf("invalid Ollama base_url")
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/api/chat"
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String(), nil
}
