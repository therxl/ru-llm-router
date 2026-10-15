package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/therxl/ru-llm-router/internal/api"
	"github.com/therxl/ru-llm-router/internal/config"
	"github.com/therxl/ru-llm-router/internal/provider/mock"
	"github.com/therxl/ru-llm-router/internal/router"
)

func main() {
	configPath := flag.String("config", "configs/example.yaml", "path to YAML configuration")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	providers := map[string]router.Provider{
		"mock": mock.New(),
	}
	rt, err := router.New(cfg, providers)
	if err != nil {
		log.Fatalf("initialize router: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("POST /v1/chat/completions", api.NewHandler(rt))

	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("ru-llm-router listening on %s", cfg.Server.Address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
