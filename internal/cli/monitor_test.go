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
	oldFactory := newMonitorService
	newMonitorService = func(config.Config) monitor.Service {
		service := monitor.NewService(cfg.MonitorStatePath, cfg.MonitorStopPath)
		service.Now = func() time.Time { return time.Unix(0, 0) }
		service.Sleep = func(time.Duration) {}
		service.StartProcess = func() error {
			return monitor.SaveState(service.StatePath, monitor.State{
				PID:         99,
				StartedAt:   time.Unix(0, 0),
				HeartbeatAt: time.Unix(0, 0),
			})
		}
		return service
	}
	t.Cleanup(func() { newMonitorService = oldFactory })

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

	oldFactory := newMonitorService
	now := time.Unix(10, 0)
	newMonitorService = func(config.Config) monitor.Service {
		service := monitor.NewService(cfg.MonitorStatePath, cfg.MonitorStopPath)
		service.Now = func() time.Time { return now }
		service.Sleep = func(d time.Duration) { now = now.Add(d) }
		service.KillProcess = func(int) error { return nil }
		return service
	}
	t.Cleanup(func() { newMonitorService = oldFactory })

	output := captureStdout(t, func() error {
		return runStopCommand(stubLogger{}, cfg)
	})

	if got := strings.TrimSpace(output); got != "Clipboard monitoring stopped." {
		t.Fatalf("stop output = %q, want %q", got, "Clipboard monitoring stopped.")
	}
}

func TestRunStatusCommandPrintsRunning(t *testing.T) {
	cfg := config.Default(t.TempDir())
	oldFactory := newMonitorService
	newMonitorService = func(config.Config) monitor.Service {
		service := monitor.NewService(cfg.MonitorStatePath, cfg.MonitorStopPath)
		service.Now = func() time.Time { return time.Unix(0, 0) }
		service.Sleep = func(time.Duration) {}
		if err := monitor.SaveState(cfg.MonitorStatePath, monitor.State{
			PID:         99,
			StartedAt:   time.Unix(0, 0),
			HeartbeatAt: time.Unix(0, 0),
		}); err != nil {
			t.Fatalf("SaveState() error = %v", err)
		}
		return service
	}
	t.Cleanup(func() { newMonitorService = oldFactory })

	output := captureStdout(t, func() error {
		return runStatusCommand(stubLogger{}, cfg)
	})

	if !strings.Contains(output, "Clipboard monitoring: running (pid 99)") {
		t.Fatalf("status output = %q, want running pid line", output)
	}
}

func TestRunMonitorCommandDelegatesToClipboardRunner(t *testing.T) {
	cfg := config.Default(t.TempDir())
	oldRunner := runClipboardMonitor
	called := false
	runClipboardMonitor = func(logger logging.Logger, gotCfg config.Config) error {
		called = true
		if gotCfg.MonitorStatePath != cfg.MonitorStatePath {
			t.Fatalf("MonitorStatePath = %q, want %q", gotCfg.MonitorStatePath, cfg.MonitorStatePath)
		}
		return nil
	}
	t.Cleanup(func() { runClipboardMonitor = oldRunner })

	if err := runMonitorCommand(stubLogger{}, cfg); err != nil {
		t.Fatalf("runMonitorCommand() error = %v", err)
	}
	if !called {
		t.Fatal("expected runClipboardMonitor to be called")
	}
}

func TestRunStartCommandReturnsServiceError(t *testing.T) {
	cfg := config.Default(t.TempDir())
	oldFactory := newMonitorService
	newMonitorService = func(config.Config) monitor.Service {
		service := monitor.NewService(cfg.MonitorStatePath, cfg.MonitorStopPath)
		service.Now = func() time.Time { return time.Unix(0, 0) }
		service.Sleep = func(time.Duration) {}
		service.StartProcess = func() error { return errors.New("boom") }
		return service
	}
	t.Cleanup(func() { newMonitorService = oldFactory })

	err := runStartCommand(stubLogger{}, cfg)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("runStartCommand() error = %v, want boom", err)
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
