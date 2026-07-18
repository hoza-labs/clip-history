package cli

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hoza-labs/clip-history/internal/config"
	"github.com/hoza-labs/clip-history/internal/monitor"
)

// stubLogger satisfies commandLogger without recording messages.
type stubLogger struct{}

// stubOutput captures printed command output.
type stubOutput struct {
	buffer bytes.Buffer
}

// stubMonitorController controls monitor command outcomes in tests.
type stubMonitorController struct {
	started bool
	stopped bool
	status  monitor.Status
	err     error
}

// Info ignores informational messages.
func (stubLogger) Info(string) {}

// Error ignores error messages.
func (stubLogger) Error(string) {}

// Debug ignores debug messages.
func (stubLogger) Debug(string) {}

// Close returns nil for compatibility with closeable logger tests.
func (stubLogger) Close() error { return nil }

// Printf records formatted output.
func (s *stubOutput) Printf(format string, args ...any) {
	s.buffer.WriteString(fmt.Sprintf(format, args...))
}

// Println records line output.
func (s *stubOutput) Println(args ...any) {
	s.buffer.WriteString(fmt.Sprintln(args...))
}

// StartMonitor returns the configured start result.
func (s stubMonitorController) StartMonitor() (bool, error) {
	return s.started, s.err
}

// StopMonitor returns the configured stop result.
func (s stubMonitorController) StopMonitor() (bool, error) {
	return s.stopped, s.err
}

// GetMonitorStatus returns the configured monitor status.
func (s stubMonitorController) GetMonitorStatus() (monitor.Status, error) {
	return s.status, s.err
}

// TestRunStartCommandPrintsStarted verifies start reports a newly started monitor.
func TestRunStartCommandPrintsStarted(t *testing.T) {
	cfg := config.Default(t.TempDir())
	oldFactory := newMonitorController
	newMonitorController = func(commandLogger, config.Config) (monitorController, error) {
		return stubMonitorController{started: true}, nil
	}
	t.Cleanup(func() { newMonitorController = oldFactory })

	output := &stubOutput{}
	if err := runStartCommand(stubLogger{}, output, cfg); err != nil {
		t.Fatalf("runStartCommand() error = %v", err)
	}

	if got := strings.TrimSpace(output.buffer.String()); got != "Clipboard monitoring started." {
		t.Fatalf("start output = %q, want %q", got, "Clipboard monitoring started.")
	}
}

// TestRunStopCommandPrintsStopped verifies stop reports a stopped monitor.
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
	newMonitorController = func(commandLogger, config.Config) (monitorController, error) {
		return stubMonitorController{stopped: true}, nil
	}
	t.Cleanup(func() { newMonitorController = oldFactory })

	output := &stubOutput{}
	if err := runStopCommand(stubLogger{}, output, cfg); err != nil {
		t.Fatalf("runStopCommand() error = %v", err)
	}

	if got := strings.TrimSpace(output.buffer.String()); got != "Clipboard monitoring stopped." {
		t.Fatalf("stop output = %q, want %q", got, "Clipboard monitoring stopped.")
	}
}

// TestRunStatusCommandPrintsRunning verifies status reports a running monitor PID.
func TestRunStatusCommandPrintsRunning(t *testing.T) {
	cfg := config.Default(t.TempDir())
	oldFactory := newMonitorController
	newMonitorController = func(commandLogger, config.Config) (monitorController, error) {
		return stubMonitorController{
			status: monitor.Status{Running: true, PID: 99},
		}, nil
	}
	t.Cleanup(func() { newMonitorController = oldFactory })

	output := &stubOutput{}
	if err := runStatusCommand(stubLogger{}, output, cfg); err != nil {
		t.Fatalf("runStatusCommand() error = %v", err)
	}

	if !strings.Contains(output.buffer.String(), "Clipboard monitoring: running (pid 99)") {
		t.Fatalf("status output = %q, want running pid line", output.buffer.String())
	}
}

// TestRunMonitorCommandDelegatesToClipboardRunner verifies the hidden command invokes the worker.
func TestRunMonitorCommandDelegatesToClipboardRunner(t *testing.T) {
	cfg := config.Default(t.TempDir())
	oldRunner := runMonitorWorker
	called := false
	runMonitorWorker = func(logger commandLogger, gotCfg config.Config) error {
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

// TestRunStartCommandReturnsServiceError verifies start propagates controller errors.
func TestRunStartCommandReturnsServiceError(t *testing.T) {
	cfg := config.Default(t.TempDir())
	oldFactory := newMonitorController
	newMonitorController = func(commandLogger, config.Config) (monitorController, error) {
		return stubMonitorController{err: errors.New("boom")}, nil
	}
	t.Cleanup(func() { newMonitorController = oldFactory })

	err := runStartCommand(stubLogger{}, &stubOutput{}, cfg)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("runStartCommand() error = %v, want boom", err)
	}
}

// TestRunStartCommandReturnsControllerConstructorError verifies start propagates constructor errors.
func TestRunStartCommandReturnsControllerConstructorError(t *testing.T) {
	cfg := config.Default(t.TempDir())
	oldFactory := newMonitorController
	newMonitorController = func(commandLogger, config.Config) (monitorController, error) {
		return nil, errors.New("bad constructor")
	}
	t.Cleanup(func() { newMonitorController = oldFactory })

	err := runStartCommand(stubLogger{}, &stubOutput{}, cfg)
	if err == nil || err.Error() != "bad constructor" {
		t.Fatalf("runStartCommand() error = %v, want bad constructor", err)
	}
}
