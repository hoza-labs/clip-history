// Package config loads, writes, normalizes, and validates the user's TOML configuration.
//
// Protocol: callers should use Load before commands that need filesystem paths; Load
// creates a default config file and required directories when they do not exist.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Config describes all persistent paths and settings used by clip-history.
type Config struct {
	Path             string `toml:"path"`
	HomeDir          string `toml:"homeDir"`
	HistoryDir       string `toml:"historyDir"`
	LogDir           string `toml:"logDir"`
	RuntimeDir       string `toml:"runtimeDir"`
	MonitorStatePath string `toml:"monitorStatePath"`
	MonitorStopPath  string `toml:"monitorStopPath"`
	LogRetentionDays int    `toml:"logRetentionDays"`
	LogTimezone      string `toml:"logTimezone"`
	SaveSensitive    bool   `toml:"saveSensitive"`
}

// Default returns a complete Config rooted under homeDir, or "." when homeDir is empty.
func Default(homeDir string) Config {
	if homeDir == "" {
		homeDir = "."
	}

	baseDir := filepath.Join(homeDir, ".clip-history")
	return Config{
		HomeDir:          homeDir,
		Path:             filepath.Join(baseDir, "config.toml"),
		HistoryDir:       filepath.Join(baseDir, "history"),
		LogDir:           filepath.Join(baseDir, "logs"),
		RuntimeDir:       filepath.Join(baseDir, "run"),
		MonitorStatePath: filepath.Join(baseDir, "run", "monitor.json"),
		MonitorStopPath:  filepath.Join(baseDir, "run", "monitor.stop"),
		LogRetentionDays: 14,
		LogTimezone:      "local",
		SaveSensitive:    false,
	}
}

// Load returns the normalized user Config, creating the config file and directories as needed.
func Load() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		homeDir = os.Getenv("HOME")
	}
	if homeDir == "" {
		return Config{}, errors.New("unable to determine home directory")
	}

	cfg := Default(homeDir)
	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0o755); err != nil {
		return Config{}, fmt.Errorf("create config directory: %w", err)
	}

	if _, err := os.Stat(cfg.Path); err == nil {
		data, err := os.ReadFile(cfg.Path)
		if err != nil {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("decode config: %w", err)
		}
	} else {
		data, err := toml.Marshal(cfg)
		if err != nil {
			return Config{}, fmt.Errorf("encode config: %w", err)
		}
		if err := os.WriteFile(cfg.Path, data, 0o644); err != nil {
			return Config{}, fmt.Errorf("write config: %w", err)
		}
	}

	cfg = normalize(cfg, homeDir)
	if err := validate(cfg); err != nil {
		return Config{}, err
	}
	if err := ensureDirectories(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// normalize fills empty fields in cfg with defaults derived from homeDir and returns the result.
func normalize(cfg Config, homeDir string) Config {
	defaults := Default(homeDir)

	if cfg.HomeDir == "" {
		cfg.HomeDir = defaults.HomeDir
	}
	if cfg.Path == "" {
		cfg.Path = defaults.Path
	}
	if cfg.HistoryDir == "" {
		cfg.HistoryDir = defaults.HistoryDir
	}
	if cfg.LogDir == "" {
		cfg.LogDir = defaults.LogDir
	}
	if cfg.RuntimeDir == "" {
		cfg.RuntimeDir = defaults.RuntimeDir
	}
	if cfg.MonitorStatePath == "" {
		cfg.MonitorStatePath = defaults.MonitorStatePath
	}
	if cfg.MonitorStopPath == "" {
		cfg.MonitorStopPath = defaults.MonitorStopPath
	}
	if cfg.LogRetentionDays <= 0 {
		cfg.LogRetentionDays = defaults.LogRetentionDays
	}
	if cfg.LogTimezone == "" {
		cfg.LogTimezone = defaults.LogTimezone
	}

	return cfg
}

// validate returns an error when cfg contains unsupported or malformed values.
func validate(cfg Config) error {
	if cfg.LogTimezone != "local" {
		if _, err := time.LoadLocation(cfg.LogTimezone); err != nil {
			return fmt.Errorf("invalid log timezone %q: %w", cfg.LogTimezone, err)
		}
	}

	return nil
}

// ensureDirectories creates the directories required by cfg.
func ensureDirectories(cfg Config) error {
	if err := os.MkdirAll(cfg.HistoryDir, 0o755); err != nil {
		return fmt.Errorf("create history directory: %w", err)
	}
	if err := os.MkdirAll(cfg.LogDir, 0o755); err != nil {
		return fmt.Errorf("create logs directory: %w", err)
	}
	if err := os.MkdirAll(cfg.RuntimeDir, 0o755); err != nil {
		return fmt.Errorf("create runtime directory: %w", err)
	}

	return nil
}
