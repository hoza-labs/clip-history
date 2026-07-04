package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hoza-labs/clip-history/internal/config"
	"github.com/hoza-labs/clip-history/internal/logging"
	"github.com/hoza-labs/clip-history/internal/monitor"
)

type stubLogger struct{}

func (stubLogger) Info(string)  {}
func (stubLogger) Error(string) {}
func (stubLogger) Debug(string) {}
func (stubLogger) Close() error { return nil }

func TestRunStartCommandPrintsStarted(t *testing.T) {
	cfg := config.Default(t.TempDir())
	oldFactory := newMonitorController
	newMonitorController = func(config.Config) (monitor.Controller, error) {
		return monitor.NewController(monitor.ControllerConfig{
			StatePath: cfg.MonitorStatePath,
			StopPath:  cfg.MonitorStopPath,
			StartProcess: func() error {
				return monitor.SaveState(cfg.MonitorStatePath, monitor.State{
					PID:         99,
					StartedAt:   time.Unix(0, 0),
					HeartbeatAt: time.Unix(0, 0),
				})
			},
			KillProcess: func(int) error { return nil },
			Now:         func() time.Time { return time.Unix(0, 0) },
			Sleep:       func(time.Duration) {},
		})
	}
	t.Cleanup(func() { newMonitorController = oldFactory })

	output := captureStdout(t, func() error {
		return runStartCommand(stubLogger{}, cfg)
	})

	if got := strings.TrimSpace(output); got != "Clipboard monitoring started." {
		t.Fatalf("start output = %q, want %q", got, "Clipboard monitoring started.")
	}
}

func TestRunStopCommandPrintsStopped(t *testing.T) {
	cfg := config.Default(t.TempDir())
	if err := monitor.SaveState(cfg.MonitorStatePath, monitor.State{
		PID:         99,
		StartedAt:   time.Unix(0, 0),
		HeartbeatAt: time.Unix(0, 0),
	}); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	oldFactory := newMonitorController
	now := time.Unix(10, 0)
	newMonitorController = func(config.Config) (monitor.Controller, error) {
		return monitor.NewController(monitor.ControllerConfig{
			StatePath:    cfg.MonitorStatePath,
			StopPath:     cfg.MonitorStopPath,
			StartProcess: func() error { return nil },
			KillProcess:  func(int) error { return nil },
			Now:          func() time.Time { return now },
			Sleep:        func(d time.Duration) { now = now.Add(d) },
		})
	}
	t.Cleanup(func() { newMonitorController = oldFactory })

	output := captureStdout(t, func() error {
		return runStopCommand(stubLogger{}, cfg)
	})

	if got := strings.TrimSpace(output); got != "Clipboard monitoring stopped." {
		t.Fatalf("stop output = %q, want %q", got, "Clipboard monitoring stopped.")
	}
}

func TestRunStatusCommandPrintsRunning(t *testing.T) {
	cfg := config.Default(t.TempDir())
	oldFactory := newMonitorController
	newMonitorController = func(config.Config) (monitor.Controller, error) {
		if err := monitor.SaveState(cfg.MonitorStatePath, monitor.State{
			PID:         99,
			StartedAt:   time.Unix(0, 0),
			HeartbeatAt: time.Unix(0, 0),
		}); err != nil {
			t.Fatalf("SaveState() error = %v", err)
		}
		return monitor.NewController(monitor.ControllerConfig{
			StatePath:    cfg.MonitorStatePath,
			StopPath:     cfg.MonitorStopPath,
			StartProcess: func() error { return nil },
			KillProcess:  func(int) error { return nil },
			Now:          func() time.Time { return time.Unix(0, 0) },
			Sleep:        func(time.Duration) {},
		})
	}
	t.Cleanup(func() { newMonitorController = oldFactory })

	output := captureStdout(t, func() error {
		return runStatusCommand(stubLogger{}, cfg)
	})

	if !strings.Contains(output, "Clipboard monitoring: running (pid 99)") {
		t.Fatalf("status output = %q, want running pid line", output)
	}
}

func TestRunMonitorCommandDelegatesToClipboardRunner(t *testing.T) {
	cfg := config.Default(t.TempDir())
	oldRunner := runMonitorWorker
	called := false
	runMonitorWorker = func(logger logging.Logger, gotCfg config.Config) error {
		called = true
		if gotCfg.MonitorStatePath != cfg.MonitorStatePath {
			t.Fatalf("MonitorStatePath = %q, want %q", gotCfg.MonitorStatePath, cfg.MonitorStatePath)
		}
		return nil
	}
	t.Cleanup(func() { runMonitorWorker = oldRunner })

	if err := runMonitorCommand(stubLogger{}, cfg); err != nil {
		t.Fatalf("runMonitorCommand() error = %v", err)
	}
	if !called {
		t.Fatal("expected runMonitorWorker to be called")
	}
}

func TestRunStartCommandReturnsServiceError(t *testing.T) {
	cfg := config.Default(t.TempDir())
	oldFactory := newMonitorController
	newMonitorController = func(config.Config) (monitor.Controller, error) {
		return monitor.NewController(monitor.ControllerConfig{
			StatePath:    cfg.MonitorStatePath,
			StopPath:     cfg.MonitorStopPath,
			StartProcess: func() error { return errors.New("boom") },
			KillProcess:  func(int) error { return nil },
			Now:          func() time.Time { return time.Unix(0, 0) },
			Sleep:        func(time.Duration) {},
		})
	}
	t.Cleanup(func() { newMonitorController = oldFactory })

	err := runStartCommand(stubLogger{}, cfg)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("runStartCommand() error = %v, want boom", err)
	}
}

func TestRunStartCommandReturnsControllerConstructorError(t *testing.T) {
	cfg := config.Default(t.TempDir())
	oldFactory := newMonitorController
	newMonitorController = func(config.Config) (monitor.Controller, error) {
		return monitor.Controller{}, errors.New("bad constructor")
	}
	t.Cleanup(func() { newMonitorController = oldFactory })

	err := runStartCommand(stubLogger{}, cfg)
	if err == nil || err.Error() != "bad constructor" {
		t.Fatalf("runStartCommand() error = %v, want bad constructor", err)
	}
}

func captureStdout(t *testing.T, fn func() error) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = oldStdout })

	err := fn()
	_ = w.Close()
	if err != nil {
		t.Fatalf("function error = %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}
