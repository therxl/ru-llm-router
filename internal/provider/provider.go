package provider

import (
	"context"
)

type Connection struct {
	Name     string
	Provider string
	Model    string
	BaseURL  string
	APIKey   string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Content string
	Model   string
}

type Provider interface {
	Chat(context.Context, Connection, ChatRequest) (ChatResponse, error)
}
