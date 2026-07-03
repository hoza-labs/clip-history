package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	DefaultHeartbeatInterval = 2 * time.Second
	DefaultHeartbeatTimeout  = 5 * time.Second
	DefaultPollInterval      = 100 * time.Millisecond
	DefaultStartTimeout      = 5 * time.Second
	DefaultStopTimeout       = 5 * time.Second
)

type State struct {
	PID         int       `json:"pid"`
	StartedAt   time.Time `json:"started_at"`
	HeartbeatAt time.Time `json:"heartbeat_at"`
}

func LoadState(path string) (State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return State{}, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, fmt.Errorf("decode monitor state: %w", err)
	}

	return state, nil
}

func SaveState(path string, state State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create monitor state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode monitor state: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write monitor state: %w", err)
	}

	return nil
}

func RemoveFile(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func IsRunning(state State, now time.Time, heartbeatTimeout time.Duration) bool {
	if state.PID <= 0 || state.HeartbeatAt.IsZero() {
		return false
	}

	return now.Sub(state.HeartbeatAt) <= heartbeatTimeout
}
