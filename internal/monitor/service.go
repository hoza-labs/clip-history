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

type ServiceConfig struct {
	StatePath        string
	StopPath         string
	StartProcess     func() error
	KillProcess      func(int) error
	HeartbeatTimeout time.Duration
	StartTimeout     time.Duration
	StopTimeout      time.Duration
	PollInterval     time.Duration
	Now              func() time.Time
	Sleep            func(time.Duration)
}

type Service struct {
	cfg ServiceConfig
}

func NewService(cfg ServiceConfig) (Service, error) {
	if cfg.StatePath == "" {
		return Service{}, errors.New("state path is required")
	}
	if cfg.StopPath == "" {
		return Service{}, errors.New("stop path is required")
	}
	if cfg.StartProcess == nil {
		return Service{}, errors.New("start process function is required")
	}
	if cfg.KillProcess == nil {
		return Service{}, errors.New("kill process function is required")
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

	return Service{cfg: cfg}, nil
}

// Start returns true if the service was started; it returns false if it was already running
// or could not be started due to error
func (s Service) Start() (bool, error) {
	status, err := s.Status()
	if err != nil {
		return false, err
	}
	if status.Running {
		return false, nil
	}

	if err := RemoveFile(s.cfg.StatePath); err != nil {
		return false, fmt.Errorf("remove stale state: %w", err)
	}
	if err := RemoveFile(s.cfg.StopPath); err != nil {
		return false, fmt.Errorf("remove stale stop signal: %w", err)
	}

	if err := s.cfg.StartProcess(); err != nil {
		return false, err
	}

	deadline := s.cfg.Now().Add(s.cfg.StartTimeout)
	for s.cfg.Now().Before(deadline) || s.cfg.Now().Equal(deadline) {
		status, err := s.Status()
		if err == nil && status.Running {
			return true, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return false, err
		}

		s.cfg.Sleep(s.cfg.PollInterval)
	}

	return false, errors.New("monitor did not become ready in time")
}

func (s Service) Stop() (bool, error) {
	state, err := LoadState(s.cfg.StatePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_ = RemoveFile(s.cfg.StopPath)
			return false, nil
		}
		return false, err
	}

	if err := os.WriteFile(s.cfg.StopPath, []byte("stop\n"), 0o644); err != nil {
		return false, fmt.Errorf("write stop signal: %w", err)
	}

	deadline := s.cfg.Now().Add(s.cfg.StopTimeout)
	for s.cfg.Now().Before(deadline) || s.cfg.Now().Equal(deadline) {
		if _, err := os.Stat(s.cfg.StatePath); errors.Is(err, os.ErrNotExist) {
			_ = RemoveFile(s.cfg.StopPath)
			return true, nil
		}

		s.cfg.Sleep(s.cfg.PollInterval)
	}

	if state.PID > 0 {
		if err := s.cfg.KillProcess(state.PID); err != nil {
			return true, err
		}
	}

	if err := RemoveFile(s.cfg.StatePath); err != nil {
		return true, fmt.Errorf("remove monitor state: %w", err)
	}
	if err := RemoveFile(s.cfg.StopPath); err != nil {
		return true, fmt.Errorf("remove stop signal: %w", err)
	}

	return true, nil
}

func (s Service) Status() (Status, error) {
	state, err := LoadState(s.cfg.StatePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Status{}, nil
		}
		return Status{}, err
	}

	if IsRunning(state, s.cfg.Now(), s.cfg.HeartbeatTimeout) {
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
