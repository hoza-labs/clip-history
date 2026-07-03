package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Path         string `json:"path"`
	HomeDir      string `json:"home_dir"`
	HistoryDir   string `json:"history_dir"`
	LogDir       string `json:"log_dir"`
	SaveSensitive bool  `json:"save_sensitive"`
}

func Default(homeDir string) Config {
	if homeDir == "" {
		homeDir = "."
	}

	baseDir := filepath.Join(homeDir, ".clip-history")
	return Config{
		HomeDir:       homeDir,
		Path:          filepath.Join(baseDir, "config.json"),
		HistoryDir:    filepath.Join(baseDir, "history"),
		LogDir:        filepath.Join(baseDir, "logs"),
		SaveSensitive: false,
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
		if err := os.MkdirAll(cfg.HistoryDir, 0o755); err != nil {
			return Config{}, fmt.Errorf("create history directory: %w", err)
		}
		if err := os.MkdirAll(cfg.LogDir, 0o755); err != nil {
			return Config{}, fmt.Errorf("create logs directory: %w", err)
		}
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return Config{}, fmt.Errorf("encode config: %w", err)
		}
		if err := os.WriteFile(cfg.Path, data, 0o644); err != nil {
			return Config{}, fmt.Errorf("write config: %w", err)
		}
	}

	return cfg, nil
}
