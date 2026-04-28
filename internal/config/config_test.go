package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cronitorio/cronitor-local/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "cronitor.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return path
}

func TestLoad_ValidConfig(t *testing.T) {
	yaml := `
api_key: "test-key"
grace_period: 5m
jobs:
  - name: backup
    schedule: "0 2 * * *"
    command: "/usr/local/bin/backup.sh"
    timeout: 30m
    notify:
      - ops@example.com
`
	path := writeTempConfig(t, yaml)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "test-key" {
		t.Errorf("expected api_key 'test-key', got %q", cfg.APIKey)
	}
	if cfg.GracePeriod != 5*time.Minute {
		t.Errorf("expected grace_period 5m, got %v", cfg.GracePeriod)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "backup" {
		t.Errorf("expected job name 'backup', got %q", cfg.Jobs[0].Name)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/cronitor.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_MissingJobName(t *testing.T) {
	yaml := `
jobs:
  - schedule: "* * * * *"
    command: "echo hello"
`
	path := writeTempConfig(t, yaml)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing job name, got nil")
	}
}

func TestLoad_MissingCommand(t *testing.T) {
	yaml := `
jobs:
  - name: test-job
    schedule: "* * * * *"
`
	path := writeTempConfig(t, yaml)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing command, got nil")
	}
}
