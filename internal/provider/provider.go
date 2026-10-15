package provider

import "context"

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
	Chat(context.Context, string, ChatRequest) (ChatResponse, error)
}
