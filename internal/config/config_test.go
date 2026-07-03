package config

import (
	"os"
	"path/filepath"
	"testing"
)

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

	if filepath.Base(cfg.Path) != "config.json" {
		t.Fatalf("unexpected config file name: %s", filepath.Base(cfg.Path))
	}
	if filepath.Base(cfg.MonitorStatePath) != "monitor.json" {
		t.Fatalf("unexpected monitor state file name: %s", filepath.Base(cfg.MonitorStatePath))
	}
}
