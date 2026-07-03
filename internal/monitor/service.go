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

type Service struct {
	StatePath        string
	StopPath         string
	HeartbeatTimeout time.Duration
	StartTimeout     time.Duration
	StopTimeout      time.Duration
	PollInterval     time.Duration
	Now              func() time.Time
	Sleep            func(time.Duration)
	StartProcess     func() error
	KillProcess      func(int) error
}

func NewService(statePath, stopPath string) Service {
	return Service{
		StatePath:        statePath,
		StopPath:         stopPath,
		HeartbeatTimeout: DefaultHeartbeatTimeout,
		StartTimeout:     DefaultStartTimeout,
		StopTimeout:      DefaultStopTimeout,
		PollInterval:     DefaultPollInterval,
		Now:              time.Now,
		Sleep:            time.Sleep,
	}
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

	if err := RemoveFile(s.StatePath); err != nil {
		return false, fmt.Errorf("remove stale state: %w", err)
	}
	if err := RemoveFile(s.StopPath); err != nil {
		return false, fmt.Errorf("remove stale stop signal: %w", err)
	}

	if s.StartProcess == nil {
		return false, errors.New("start process function is required")
	}
	if err := s.StartProcess(); err != nil {
		return false, err
	}

	deadline := s.Now().Add(s.StartTimeout)
	for s.Now().Before(deadline) || s.Now().Equal(deadline) {
		status, err := s.Status()
		if err == nil && status.Running {
			return true, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return false, err
		}

		s.Sleep(s.PollInterval)
	}

	return false, errors.New("monitor did not become ready in time")
}

func (s Service) Stop() (bool, error) {
	state, err := LoadState(s.StatePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_ = RemoveFile(s.StopPath)
			return false, nil
		}
		return false, err
	}

	if err := os.WriteFile(s.StopPath, []byte("stop\n"), 0o644); err != nil {
		return false, fmt.Errorf("write stop signal: %w", err)
	}

	deadline := s.Now().Add(s.StopTimeout)
	for s.Now().Before(deadline) || s.Now().Equal(deadline) {
		if _, err := os.Stat(s.StatePath); errors.Is(err, os.ErrNotExist) {
			_ = RemoveFile(s.StopPath)
			return true, nil
		}

		s.Sleep(s.PollInterval)
	}

	if s.KillProcess != nil && state.PID > 0 {
		if err := s.KillProcess(state.PID); err != nil {
			return true, err
		}
	}

	if err := RemoveFile(s.StatePath); err != nil {
		return true, fmt.Errorf("remove monitor state: %w", err)
	}
	if err := RemoveFile(s.StopPath); err != nil {
		return true, fmt.Errorf("remove stop signal: %w", err)
	}

	return true, nil
}

func (s Service) Status() (Status, error) {
	state, err := LoadState(s.StatePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Status{}, nil
		}
		return Status{}, err
	}

	if IsRunning(state, s.Now(), s.HeartbeatTimeout) {
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
