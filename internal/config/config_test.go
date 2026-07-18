package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

// TestLoadCreatesDefaultConfigInHomeDir verifies Load initializes config and directories.
func TestLoadCreatesDefaultConfigInHomeDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Path == "" {
		t.Fatal("expected config path to be populated")
	}
	if _, err := os.Stat(cfg.Path); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
	if _, err := os.Stat(cfg.HistoryDir); err != nil {
		t.Fatalf("expected history directory to exist: %v", err)
	}
	if _, err := os.Stat(cfg.LogDir); err != nil {
		t.Fatalf("expected log directory to exist: %v", err)
	}
	if _, err := os.Stat(cfg.RuntimeDir); err != nil {
		t.Fatalf("expected runtime directory to exist: %v", err)
	}

	if filepath.Base(cfg.Path) != "config.toml" {
		t.Fatalf("unexpected config file name: %s", filepath.Base(cfg.Path))
	}
	data, err := os.ReadFile(cfg.Path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	for _, key := range []string{"homeDir =", "historyDir =", "logDir =", "runtimeDir =", "monitorStatePath =", "monitorStopPath =", "logRetentionDays =", "logTimezone =", "saveSensitive ="} {
		if !strings.Contains(string(data), key) {
			t.Fatalf("config file missing TOML key %q:\n%s", key, string(data))
		}
	}
	if filepath.Base(cfg.MonitorStatePath) != "monitor.json" {
		t.Fatalf("unexpected monitor state file name: %s", filepath.Base(cfg.MonitorStatePath))
	}
	if cfg.LogRetentionDays != 14 {
		t.Fatalf("log retention days = %d, want 14", cfg.LogRetentionDays)
	}
	if cfg.LogTimezone != "local" {
		t.Fatalf("log timezone = %q, want local", cfg.LogTimezone)
	}
}

// TestLoadRejectsInvalidLogTimezone verifies invalid IANA time zones are rejected.
func TestLoadRejectsInvalidLogTimezone(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

	cfg := Default(dir)
	cfg.LogTimezone = "Not/AZone"
	data, err := toml.Marshal(cfg)
	if err != nil {
		t.Fatalf("encode config: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.WriteFile(cfg.Path, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := Load(); err == nil {
		t.Fatal("expected Load to reject invalid log timezone")
	}
}
