// Package monitor coordinates the background clipboard monitor worker.
//
// Protocol: callers create a Controller with NewController, then use StartMonitor,
// StopMonitor, and GetMonitorStatus against the state and stop-signal files. The
// hidden CLI "monitor" command calls RunWorker directly; RunWorker owns the worker
// loop, periodically writes heartbeat state, and exits when the stop file appears.
package monitor

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// Status describes whether the monitor worker appears active.
type Status struct {
	Running bool
	PID     int
	Stale   bool
}

// ControllerConfig provides the required file paths and logger plus optional timing/test hooks for Controller.
type ControllerConfig struct {
	StatePath        string
	StopPath         string
	HeartbeatTimeout time.Duration
	StartTimeout     time.Duration
	StopTimeout      time.Duration
	PollInterval     time.Duration
	Now              func() time.Time
	Sleep            func(time.Duration)
	Logger           ControllerLogger
}

// ControllerLogger records lifecycle decisions made by Controller.
type ControllerLogger interface {
	// Info records an informational lifecycle message.
	Info(string)
}

// Controller controls the separate long-running clipboard monitor process.
type Controller struct {
	cfg ControllerConfig
}

// NewController validates cfg and returns a Controller; StatePath, StopPath, and Logger are required.
func NewController(cfg ControllerConfig) (Controller, error) {
	if cfg.StatePath == "" {
		return Controller{}, errors.New("state path is required")
	}
	if cfg.StopPath == "" {
		return Controller{}, errors.New("stop path is required")
	}
	if cfg.Logger == nil {
		return Controller{}, errors.New("logger is required")
	}
	if cfg.HeartbeatTimeout == 0 {
		cfg.HeartbeatTimeout = DefaultHeartbeatTimeout
	}
	if cfg.StartTimeout == 0 {
		cfg.StartTimeout = DefaultStartTimeout
	}
	if cfg.StopTimeout == 0 {
		cfg.StopTimeout = DefaultStopTimeout
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = DefaultPollInterval
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Sleep == nil {
		cfg.Sleep = time.Sleep
	}
	return Controller{cfg: cfg}, nil
}

// StartMonitor starts the worker and returns true when a new process reaches a healthy state.
func (mc Controller) StartMonitor() (bool, error) {
	status, err := mc.GetMonitorStatus()
	if err != nil {
		return false, err
	}
	if status.Running {
		mc.cfg.Logger.Info(fmt.Sprintf("monitor already running with pid %d; start not needed", status.PID))
		return false, nil
	}

	if err := RemoveFile(mc.cfg.StatePath); err != nil {
		return false, fmt.Errorf("remove stale state: %w", err)
	}
	if err := RemoveFile(mc.cfg.StopPath); err != nil {
		return false, fmt.Errorf("remove stale stop signal: %w", err)
	}

	if err := startMonitorProcess(); err != nil {
		return false, err
	}
	mc.cfg.Logger.Info("monitor process started; waiting for healthy state")

	deadline := mc.cfg.Now().Add(mc.cfg.StartTimeout)
	for mc.cfg.Now().Before(deadline) || mc.cfg.Now().Equal(deadline) {
		status, err := mc.GetMonitorStatus()
		if err == nil && status.Running {
			mc.cfg.Logger.Info(fmt.Sprintf("started monitor is healthy and running with pid %d", status.PID))
			return true, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return false, err
		}

		mc.cfg.Sleep(mc.cfg.PollInterval)
	}

	return false, errors.New("monitor did not become ready in time")
}

// StopMonitor signals the worker to stop and returns true when a running worker was targeted.
func (mc Controller) StopMonitor() (bool, error) {
	state, err := LoadState(mc.cfg.StatePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_ = RemoveFile(mc.cfg.StopPath)
			mc.cfg.Logger.Info("monitor state file not found; stop not needed")
			return false, nil
		}
		return false, err
	}

	if err := os.WriteFile(mc.cfg.StopPath, []byte("stop\n"), 0o644); err != nil {
		return false, fmt.Errorf("write stop signal: %w", err)
	}

	deadline := mc.cfg.Now().Add(mc.cfg.StopTimeout)
	for mc.cfg.Now().Before(deadline) || mc.cfg.Now().Equal(deadline) {
		if _, err := os.Stat(mc.cfg.StatePath); errors.Is(err, os.ErrNotExist) {
			_ = RemoveFile(mc.cfg.StopPath)
			mc.cfg.Logger.Info("monitor state file disappeared after stop signal; monitor stopped cleanly")
			return true, nil
		}

		mc.cfg.Sleep(mc.cfg.PollInterval)
	}

	if state.PID > 0 {
		mc.cfg.Logger.Info(fmt.Sprintf("monitor state file remained after stop timeout; issuing kill command for pid %d", state.PID))
		if err := killMonitorProcess(state.PID); err != nil {
			return true, err
		}
	}

	if err := RemoveFile(mc.cfg.StatePath); err != nil {
		return true, fmt.Errorf("remove monitor state: %w", err)
	}
	if err := RemoveFile(mc.cfg.StopPath); err != nil {
		return true, fmt.Errorf("remove stop signal: %w", err)
	}

	return true, nil
}

// GetMonitorStatus returns running, stale, or stopped status from the worker state file.
func (mc Controller) GetMonitorStatus() (Status, error) {
	state, err := LoadState(mc.cfg.StatePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Status{}, nil
		}
		return Status{}, err
	}

	if IsRunning(state, mc.cfg.Now(), mc.cfg.HeartbeatTimeout) {
		return Status{
			Running: true,
			PID:     state.PID,
		}, nil
	}

	return Status{
		PID:   state.PID,
		Stale: true,
	}, nil
}
