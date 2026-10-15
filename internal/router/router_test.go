package router

import (
	"context"
	"strings"
	"testing"

	"github.com/therxl/ru-llm-router/internal/config"
	"github.com/therxl/ru-llm-router/internal/provider"
	"github.com/therxl/ru-llm-router/internal/provider/mock"
)

func TestRoutesByConnectionNameAndModel(t *testing.T) {
	cfg := config.Config{Connections: []config.ConnectionConfig{{Name: "friendly-name", Provider: "mock", Model: "underlying-model"}}}
	rt, err := New(cfg, map[string]Provider{"mock": mock.New()})
	if err != nil {
		t.Fatal(err)
	}
	for _, model := range []string{"friendly-name", "underlying-model"} {
		got, err := rt.Chat(context.Background(), provider.ChatRequest{Model: model, Messages: []provider.Message{{Role: "user", Content: "hello"}}})
		if err != nil {
			t.Fatal(err)
		}
		if got.Model != "underlying-model" || !strings.Contains(got.Content, "hello") {
			t.Fatalf("unexpected response for %q: %#v", model, got)
		}
	}
}

func TestUnknownModel(t *testing.T) {
	rt, err := New(config.Config{Connections: []config.ConnectionConfig{{Name: "known", Provider: "mock", Model: "model"}}}, map[string]Provider{"mock": mock.New()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = rt.Chat(context.Background(), provider.ChatRequest{Model: "missing"})
	if err == nil {
		t.Fatal("expected error for unknown model")
	}
}
