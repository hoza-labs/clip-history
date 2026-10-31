package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

// TestLoadCreatesDefaultConfigInHomeDir verifies Load initializes config and directories.
func TestLoadCreatesDefaultConfigInHomeDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

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
	var decoded Config
	if err := toml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode commented config: %v", err)
	}
	if decoded != cfg {
		t.Fatalf("decoded config = %+v, want %+v", decoded, cfg)
	}
	typ := reflect.TypeOf(cfg)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != typ.NumField() {
		t.Fatalf("config has %d lines, want %d", len(lines), typ.NumField())
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		description := field.Tag.Get("description")
		if description == "" {
			t.Fatalf("field %s has no description", field.Name)
		}
		if strings.TrimSpace(strings.SplitN(lines[i], "=", 2)[0]) != field.Tag.Get("toml") || !strings.HasSuffix(lines[i], " # "+description) {
			t.Errorf("field %s missing inline description: %s", field.Name, lines[i])
		}
	}
	if _, err := Load(); err != nil {
		t.Fatalf("reload commented config: %v", err)
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

// TestMarshalWithCommentsAlignment verifies aligned columns preserve quoted string contents.
func TestMarshalWithCommentsAlignment(t *testing.T) {
	cfg := Default("home")
	cfg.HistoryDir = "  clipboard = items # saved  "
	cfg.LogDir = "C:\\logs\\\"quoted\""
	cfg.RuntimeDir = "???\tfiles\nnext line"
	data, err := marshalWithComments(cfg)
	if err != nil {
		t.Fatalf("encode config: %v", err)
	}
	var decoded Config
	if err := toml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode config: %v", err)
	}
	if decoded != cfg {
		t.Fatalf("decoded config = %+v, want %+v", decoded, cfg)
	}
	equalsColumn, commentColumn := -1, -1
	for _, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
		equals := len([]rune(line[:strings.Index(line, " = ")+1]))
		comment := len([]rune(line[:strings.LastIndex(line, " # ")+1]))
		if equalsColumn == -1 {
			equalsColumn, commentColumn = equals, comment
		}
		if equals != equalsColumn || comment != commentColumn {
			t.Errorf("misaligned columns in %q: got (%d, %d), want (%d, %d)", line, equals, comment, equalsColumn, commentColumn)
		}
	}
}
