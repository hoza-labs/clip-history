package monitor

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestControllerStartWaitsForHealthyState verifies StartMonitor waits for fresh worker state.
func TestControllerStartWaitsForHealthyState(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	statePath := filepath.Join(dir, "monitor.json")
	stopPath := filepath.Join(dir, "monitor.stop")
	oldStart := startMonitorProcess
	oldKill := killMonitorProcess
	startMonitorProcess = func() error {
		return SaveState(statePath, State{
			PID:         42,
			StartedAt:   now,
			HeartbeatAt: now,
		})
	}
	killMonitorProcess = func(int) error { return nil }
	t.Cleanup(func() {
		startMonitorProcess = oldStart
		killMonitorProcess = oldKill
	})
	controller, err := NewController(ControllerConfig{
		StatePath: statePath,
		StopPath:  stopPath,
		Now:       func() time.Time { return now },
		Sleep: func(d time.Duration) {
			now = now.Add(d)
		},
		Logger: &recordingControllerLogger{},
	})
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}

	started, err := controller.StartMonitor()
	if err != nil {
		t.Fatalf("StartMonitor() error = %v", err)
	}
	if !started {
		t.Fatal("expected service to start monitoring")
	}
}

// TestControllerStopRemovesStateWhenWorkerDoesNotExit verifies StopMonitor force-cleans stale state.
func TestControllerStopRemovesStateWhenWorkerDoesNotExit(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	statePath := filepath.Join(dir, "monitor.json")
	stopPath := filepath.Join(dir, "monitor.stop")
	killedPID := 0
	oldStart := startMonitorProcess
	oldKill := killMonitorProcess
	startMonitorProcess = func() error { return nil }
	killMonitorProcess = func(pid int) error {
		killedPID = pid
		return nil
	}
	t.Cleanup(func() {
		startMonitorProcess = oldStart
		killMonitorProcess = oldKill
	})

	logger := &recordingControllerLogger{}
	controller, err := NewController(ControllerConfig{
		StatePath: statePath,
		StopPath:  stopPath,
		Now:       func() time.Time { return now },
		Sleep: func(d time.Duration) {
			now = now.Add(d)
		},
		Logger: logger,
	})
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}
	if err := SaveState(statePath, State{
		PID:         42,
		StartedAt:   now,
		HeartbeatAt: now,
	}); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	stopped, err := controller.StopMonitor()
	if err != nil {
		t.Fatalf("StopMonitor() error = %v", err)
	}
	if !stopped {
		t.Fatal("expected service to report it stopped a running monitor")
	}
	if killedPID != 42 {
		t.Fatalf("KillProcess() pid = %d, want %d", killedPID, 42)
	}
	if !strings.Contains(strings.Join(logger.messages, "\n"), "issuing kill command for pid 42") {
		t.Fatalf("messages = %#v, want kill decision", logger.messages)
	}
}

// TestControllerStatusReportsStaleWorker verifies old heartbeats are reported as stale.
func TestControllerStatusReportsStaleWorker(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 7, 3, 12, 0, 10, 0, time.UTC)
	oldStart := startMonitorProcess
	oldKill := killMonitorProcess
	startMonitorProcess = func() error { return nil }
	killMonitorProcess = func(int) error { return nil }
	t.Cleanup(func() {
		startMonitorProcess = oldStart
		killMonitorProcess = oldKill
	})
	controller, err := NewController(ControllerConfig{
		StatePath: filepath.Join(dir, "monitor.json"),
		StopPath:  filepath.Join(dir, "monitor.stop"),
		Now:       func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}
	if err := SaveState(filepath.Join(dir, "monitor.json"), State{
		PID:         42,
		StartedAt:   now.Add(-10 * time.Second),
		HeartbeatAt: now.Add(-10 * time.Second),
	}); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	status, err := controller.GetMonitorStatus()
	if err != nil {
		t.Fatalf("GetMonitorStatus() error = %v", err)
	}
	if status.Running {
		t.Fatal("expected stale worker to report not running")
	}
	if !status.Stale {
		t.Fatal("expected stale worker to be marked stale")
	}
}

// TestControllerStartReturnsErrorWhenSpawnerFails verifies spawn failures are returned.
func TestControllerStartReturnsErrorWhenSpawnerFails(t *testing.T) {
	oldStart := startMonitorProcess
	oldKill := killMonitorProcess
	startMonitorProcess = func() error { return errors.New("spawn failed") }
	killMonitorProcess = func(int) error { return nil }
	t.Cleanup(func() {
		startMonitorProcess = oldStart
		killMonitorProcess = oldKill
	})
	controller, err := NewController(ControllerConfig{
		StatePath: "monitor.json",
		StopPath:  "monitor.stop",
		Now:       func() time.Time { return time.Unix(0, 0) },
		Sleep:     func(time.Duration) {},
	})
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}

	started, err := controller.StartMonitor()
	if err == nil {
		t.Fatal("expected StartMonitor() to fail")
	}
	if started {
		t.Fatal("expected Start() to report not started")
	}
}

// TestNewControllerRequiresAllMinimalDependencies verifies required controller paths are enforced.
func TestNewControllerRequiresAllMinimalDependencies(t *testing.T) {
	_, err := NewController(ControllerConfig{})
	if err == nil {
		t.Fatal("expected NewController() to fail")
	}
}

// recordingControllerLogger captures controller lifecycle messages.
type recordingControllerLogger struct {
	messages []string
}

// Info records message.
func (l *recordingControllerLogger) Info(message string) {
	l.messages = append(l.messages, message)
}
