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
	// DefaultHeartbeatInterval is how often the worker refreshes its state file.
	DefaultHeartbeatInterval = 2 * time.Second
	// DefaultHeartbeatTimeout is how long a heartbeat remains fresh enough to count as running.
	DefaultHeartbeatTimeout = 5 * time.Second
	// DefaultPollInterval is how often the controller checks worker start/stop progress.
	DefaultPollInterval = 100 * time.Millisecond
	// DefaultStartTimeout is how long StartMonitor waits for a healthy worker.
	DefaultStartTimeout = 5 * time.Second
	// DefaultStopTimeout is how long StopMonitor waits before force-killing a worker.
	DefaultStopTimeout = 5 * time.Second
)

// State is the JSON heartbeat record written by the monitor worker.
type State struct {
	PID         int       `json:"pid"`
	StartedAt   time.Time `json:"started_at"`
	HeartbeatAt time.Time `json:"heartbeat_at"`
}

// LoadState reads path and returns the decoded worker State.
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

// SaveState writes state as JSON to path, creating the parent directory when needed.
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

// RemoveFile removes path and returns nil when the file is already absent.
func RemoveFile(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

// IsRunning reports whether state has a valid PID and a heartbeat within heartbeatTimeout of now.
func IsRunning(state State, now time.Time, heartbeatTimeout time.Duration) bool {
	if state.PID <= 0 || state.HeartbeatAt.IsZero() {
		return false
	}

	return now.Sub(state.HeartbeatAt) <= heartbeatTimeout
}
