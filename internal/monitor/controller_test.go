package monitor

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

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

	controller, err := NewController(ControllerConfig{
		StatePath: statePath,
		StopPath:  stopPath,
		Now:       func() time.Time { return now },
		Sleep: func(d time.Duration) {
			now = now.Add(d)
		},
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
}

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

func TestNewControllerRequiresAllMinimalDependencies(t *testing.T) {
	_, err := NewController(ControllerConfig{})
	if err == nil {
		t.Fatal("expected NewController() to fail")
	}
}
