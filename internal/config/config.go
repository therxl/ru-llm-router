package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server      ServerConfig       `yaml:"server"`
	Connections []ConnectionConfig `yaml:"connections"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
}

type ConnectionConfig struct {
	Name       string `yaml:"name"`
	Provider   string `yaml:"provider"`
	Model      string `yaml:"model"`
	BaseURL    string `yaml:"base_url,omitempty"`
	APIKeyEnv  string `yaml:"api_key_env,omitempty"`
	APIKey     string `yaml:"-"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse YAML: %w", err)
	}
	if strings.TrimSpace(cfg.Server.Address) == "" {
		cfg.Server.Address = "127.0.0.1:8080"
	}
	if len(cfg.Connections) == 0 {
		return Config{}, fmt.Errorf("at least one connection must be configured")
	}

	seen := make(map[string]struct{}, len(cfg.Connections))
	for i := range cfg.Connections {
		conn := &cfg.Connections[i]
		conn.Name = strings.TrimSpace(conn.Name)
		conn.Provider = strings.ToLower(strings.TrimSpace(conn.Provider))
		conn.Model = strings.TrimSpace(conn.Model)
		if conn.Name == "" || conn.Model == "" || conn.Provider == "" {
			return Config{}, fmt.Errorf("connection at index %d requires name, provider, and model", i)
		}
		if _, ok := seen[conn.Name]; ok {
			return Config{}, fmt.Errorf("duplicate connection name %q", conn.Name)
		}
		seen[conn.Name] = struct{}{}

		if conn.APIKeyEnv != "" {
			conn.APIKey = os.Getenv(conn.APIKeyEnv)
			if conn.APIKey == "" {
				return Config{}, fmt.Errorf("connection %q requires environment variable %s", conn.Name, conn.APIKeyEnv)
			}
		}
	}
	return cfg, nil
}
