package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Path             string `json:"path"`
	HomeDir          string `json:"home_dir"`
	HistoryDir       string `json:"history_dir"`
	LogDir           string `json:"log_dir"`
	RuntimeDir       string `json:"runtime_dir"`
	MonitorStatePath string `json:"monitor_state_path"`
	MonitorStopPath  string `json:"monitor_stop_path"`
	SaveSensitive    bool   `json:"save_sensitive"`
}

func Default(homeDir string) Config {
	if homeDir == "" {
		homeDir = "."
	}

	baseDir := filepath.Join(homeDir, ".clip-history")
	return Config{
		HomeDir:          homeDir,
		Path:             filepath.Join(baseDir, "config.json"),
		HistoryDir:       filepath.Join(baseDir, "history"),
		LogDir:           filepath.Join(baseDir, "logs"),
		RuntimeDir:       filepath.Join(baseDir, "run"),
		MonitorStatePath: filepath.Join(baseDir, "run", "monitor.json"),
		MonitorStopPath:  filepath.Join(baseDir, "run", "monitor.stop"),
		SaveSensitive:    false,
	}
}

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
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("decode config: %w", err)
		}
	} else {
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return Config{}, fmt.Errorf("encode config: %w", err)
		}
		if err := os.WriteFile(cfg.Path, data, 0o644); err != nil {
			return Config{}, fmt.Errorf("write config: %w", err)
		}
	}

	cfg = normalize(cfg, homeDir)
	if err := ensureDirectories(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

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

	return cfg
}

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
