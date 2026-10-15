package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadDefaultsAddress(t *testing.T) {
	path := writeConfig(t, "connections:\n  - name: test\n    provider: mock\n    model: demo\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Address != "127.0.0.1:8080" {
		t.Fatalf("unexpected address: %q", cfg.Server.Address)
	}
}

func TestLoadRequiresSecretFromEnvironment(t *testing.T) {
	t.Setenv("ROUTER_TEST_API_KEY", "")
	path := writeConfig(t, "connections:\n  - name: test\n    provider: mock\n    model: demo\n    api_key_env: ROUTER_TEST_API_KEY\n")
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "ROUTER_TEST_API_KEY") {
		t.Fatalf("expected missing environment variable error, got %v", err)
	}
}

func TestLoadRejectsDuplicateNames(t *testing.T) {
	path := writeConfig(t, "connections:\n  - name: test\n    provider: mock\n    model: demo\n  - name: test\n    provider: mock\n    model: other\n")
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "duplicate connection") {
		t.Fatalf("expected duplicate connection error, got %v", err)
	}
}
