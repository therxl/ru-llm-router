package mock

import (
	"context"
	"fmt"

	"github.com/therxl/ru-llm-router/internal/provider"
)

type Provider struct{}

func New() *Provider { return &Provider{} }

func (p *Provider) Chat(_ context.Context, connection provider.Connection, req provider.ChatRequest) (provider.ChatResponse, error) {
	var last string
	for _, message := range req.Messages {
		if message.Role == "user" {
			last = message.Content
		}
	}
	if last == "" {
		last = "(no user message)"
	}
	return provider.ChatResponse{
		Model:   connection.Model,
		Content: fmt.Sprintf("Mock response from %s: received your message: %s", connection.Model, last),
	}, nil
}
