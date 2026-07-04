package monitor

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestServiceStartWaitsForHealthyState(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	statePath := filepath.Join(dir, "monitor.json")
	stopPath := filepath.Join(dir, "monitor.stop")
	service, err := NewService(ServiceConfig{
		StatePath: statePath,
		StopPath:  stopPath,
		StartProcess: func() error {
			return SaveState(statePath, State{
				PID:         42,
				StartedAt:   now,
				HeartbeatAt: now,
			})
		},
		KillProcess: func(int) error { return nil },
		Now:         func() time.Time { return now },
		Sleep: func(d time.Duration) {
			now = now.Add(d)
		},
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	started, err := service.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !started {
		t.Fatal("expected service to start monitoring")
	}
}

func TestServiceStopRemovesStateWhenMonitorDoesNotExit(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	statePath := filepath.Join(dir, "monitor.json")
	stopPath := filepath.Join(dir, "monitor.stop")
	killedPID := 0

	service, err := NewService(ServiceConfig{
		StatePath:    statePath,
		StopPath:     stopPath,
		StartProcess: func() error { return nil },
		KillProcess: func(pid int) error {
			killedPID = pid
			return nil
		},
		Now: func() time.Time { return now },
		Sleep: func(d time.Duration) {
			now = now.Add(d)
		},
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if err := SaveState(service.StatePath, State{
		PID:         42,
		StartedAt:   now,
		HeartbeatAt: now,
	}); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	stopped, err := service.Stop()
	if err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if !stopped {
		t.Fatal("expected service to report it stopped a running monitor")
	}
	if killedPID != 42 {
		t.Fatalf("KillProcess() pid = %d, want %d", killedPID, 42)
	}
}

func TestServiceStatusReportsStaleMonitor(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 7, 3, 12, 0, 10, 0, time.UTC)
	service, err := NewService(ServiceConfig{
		StatePath:    filepath.Join(dir, "monitor.json"),
		StopPath:     filepath.Join(dir, "monitor.stop"),
		StartProcess: func() error { return nil },
		KillProcess:  func(int) error { return nil },
		Now:          func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if err := SaveState(service.StatePath, State{
		PID:         42,
		StartedAt:   now.Add(-10 * time.Second),
		HeartbeatAt: now.Add(-10 * time.Second),
	}); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	status, err := service.Status()
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Running {
		t.Fatal("expected stale monitor to report not running")
	}
	if !status.Stale {
		t.Fatal("expected stale monitor to be marked stale")
	}
}

func TestServiceStartReturnsErrorWhenSpawnerFails(t *testing.T) {
	service, err := NewService(ServiceConfig{
		StatePath:    "monitor.json",
		StopPath:     "monitor.stop",
		StartProcess: func() error { return errors.New("spawn failed") },
		KillProcess:  func(int) error { return nil },
		Now:          func() time.Time { return time.Unix(0, 0) },
		Sleep:        func(time.Duration) {},
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	started, err := service.Start()
	if err == nil {
		t.Fatal("expected Start() to fail")
	}
	if started {
		t.Fatal("expected Start() to report not started")
	}
}

func TestNewServiceRequiresAllMinimalDependencies(t *testing.T) {
	_, err := NewService(ServiceConfig{})
	if err == nil {
		t.Fatal("expected NewService() to fail")
	}
}
