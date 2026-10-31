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
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pelletier/go-toml/v2"
)

// Config describes all persistent paths and settings used by clip-history.
type Config struct {
	Path             string `toml:"path" description:"Path to the configuration file."`
	HomeDir          string `toml:"homeDir" description:"User home directory used as the configuration base."`
	HistoryDir       string `toml:"historyDir" description:"Directory that holds actual clipboard items."`
	LogDir           string `toml:"logDir" description:"Directory for log files."`
	RuntimeDir       string `toml:"runtimeDir" description:"Directory for monitor runtime files."`
	MonitorStatePath string `toml:"monitorStatePath" description:"Path to the monitor process state file."`
	MonitorStopPath  string `toml:"monitorStopPath" description:"Path to the file that signals the monitor to stop."`
	LogRetentionDays int    `toml:"logRetentionDays" description:"Number of days to retain log files."`
	LogTimezone      string `toml:"logTimezone" description:"Time zone for log timestamps and rotation (local or an IANA name)."`
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
		data, err := marshalWithComments(cfg)
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

// marshalWithComments encodes cfg with aligned keys, values, and field descriptions.
func marshalWithComments(cfg Config) ([]byte, error) {
	data, err := toml.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	descriptions := make(map[string]string)
	typ := reflect.TypeOf(cfg)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		descriptions[field.Tag.Get("toml")] = field.Tag.Get("description")
	}
	lines := strings.Split(string(data), "\n")
	keyWidth, valueWidth := 0, 0
	for _, line := range lines {
		if key, value, ok := strings.Cut(line, " = "); ok {
			keyWidth = max(keyWidth, utf8.RuneCountInString(key))
			valueWidth = max(valueWidth, utf8.RuneCountInString(value))
		}
	}
	for i, line := range lines {
		key, value, ok := strings.Cut(line, " = ")
		if !ok {
			continue
		}
		lines[i] = fmt.Sprintf("%-*s = %s", keyWidth, key, value)
		if description := descriptions[key]; description != "" {
			lines[i] = fmt.Sprintf("%-*s = %-*s # %s", keyWidth, key, valueWidth, value, description)
		}
	}
	return []byte(strings.Join(lines, "\n")), nil
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
