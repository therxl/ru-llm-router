package router

import (
	"context"
	"fmt"

	"github.com/therxl/ru-llm-router/internal/config"
	"github.com/therxl/ru-llm-router/internal/provider"
)

type Provider = provider.Provider
type ChatRequest = provider.ChatRequest
type ChatResponse = provider.ChatResponse

type Router struct {
	connections map[string]config.ConnectionConfig
	providers   map[string]Provider
}

func New(cfg config.Config, providers map[string]Provider) (*Router, error) {
	connections := make(map[string]config.ConnectionConfig, len(cfg.Connections))
	for _, connection := range cfg.Connections {
		if _, ok := providers[connection.Provider]; !ok {
			return nil, fmt.Errorf("connection %q uses unsupported provider %q", connection.Name, connection.Provider)
		}
		if _, exists := connections[connection.Name]; exists {
			return nil, fmt.Errorf("duplicate connection name %q", connection.Name)
		}
		connections[connection.Name] = connection
		if _, exists := connections[connection.Model]; !exists {
			connections[connection.Model] = connection
		}
	}
	return &Router{connections: connections, providers: providers}, nil
}

func (r *Router) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	connection, ok := r.connections[req.Model]
	if !ok {
		return ChatResponse{}, fmt.Errorf("unknown model or connection %q", req.Model)
	}
	adapter := r.providers[connection.Provider]
	req.Model = connection.Model
	response, err := adapter.Chat(ctx, provider.Connection{
		Name:     connection.Name,
		Provider: connection.Provider,
		Model:    connection.Model,
		BaseURL:  connection.BaseURL,
		APIKey:   connection.APIKey,
	}, req)
	if err != nil {
		return ChatResponse{}, err
	}
	return response, nil
}
