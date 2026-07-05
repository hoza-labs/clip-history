package monitor

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type Status struct {
	Running bool
	PID     int
	Stale   bool
}

type ControllerConfig struct {
	StatePath        string
	StopPath         string
	HeartbeatTimeout time.Duration
	StartTimeout     time.Duration
	StopTimeout      time.Duration
	PollInterval     time.Duration
	Now              func() time.Time
	Sleep            func(time.Duration)
}

// A Controller is the thing that controls the clipboard monitor, which is a separate process,
// running in the background. It's a separate, long-running instance of this same program.
// The receiver variable 'mc' in this file stands for "monitor controller".
type Controller struct {
	cfg ControllerConfig
}

func NewController(cfg ControllerConfig) (Controller, error) {
	if cfg.StatePath == "" {
		return Controller{}, errors.New("state path is required")
	}
	if cfg.StopPath == "" {
		return Controller{}, errors.New("stop path is required")
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

// StartMonitor returns true if monitoring was started; it returns false if it was
// already running or could not be started due to error.
func (mc Controller) StartMonitor() (bool, error) {
	status, err := mc.GetMonitorStatus()
	if err != nil {
		return false, err
	}
	if status.Running {
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

	deadline := mc.cfg.Now().Add(mc.cfg.StartTimeout)
	for mc.cfg.Now().Before(deadline) || mc.cfg.Now().Equal(deadline) {
		status, err := mc.GetMonitorStatus()
		if err == nil && status.Running {
			return true, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return false, err
		}

		mc.cfg.Sleep(mc.cfg.PollInterval)
	}

	return false, errors.New("monitor did not become ready in time")
}

func (mc Controller) StopMonitor() (bool, error) {
	state, err := LoadState(mc.cfg.StatePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_ = RemoveFile(mc.cfg.StopPath)
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
			return true, nil
		}

		mc.cfg.Sleep(mc.cfg.PollInterval)
	}

	if state.PID > 0 {
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
