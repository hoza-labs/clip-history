package cli

import (
	"os"

	"github.com/hoza-labs/clip-history/internal/config"
	"github.com/hoza-labs/clip-history/internal/monitor"
)

// runStartCommand starts background monitoring and reports whether it was already running.
func runStartCommand(logger commandLogger, output commandOutput, cfg config.Config) error {
	logger.Info("running start command")

	controller, err := newMonitorController(logger, cfg)
	if err != nil {
		return err
	}
	started, err := controller.StartMonitor()
	if err != nil {
		return err
	}
	if !started {
		msg := "Clipboard monitoring is already running."
		output.Println(msg)
		return nil
	}

	output.Println("Clipboard monitoring started.")
	return nil
}

// runStopCommand stops background monitoring and reports whether no worker was running.
func runStopCommand(logger commandLogger, output commandOutput, cfg config.Config) error {
	logger.Info("running stop command")

	controller, err := newMonitorController(logger, cfg)
	if err != nil {
		return err
	}
	stopped, err := controller.StopMonitor()
	if err != nil {
		return err
	}
	if !stopped {
		output.Println("Clipboard monitoring is not running.")
		return nil
	}

	output.Println("Clipboard monitoring stopped.")
	return nil
}

// runStatusCommand prints the history directory and monitor status, clearing stale state when found.
func runStatusCommand(logger commandLogger, output commandOutput, cfg config.Config) error {
	logger.Info("running status command")

	controller, err := newMonitorController(logger, cfg)
	if err != nil {
		return err
	}
	status, err := controller.GetMonitorStatus()
	if err != nil {
		return err
	}

	output.Printf("Clipboard history directory: %s\n", cfg.HistoryDir)
	if status.Running {
		output.Printf("Clipboard monitoring: running (pid %d)\n", status.PID)
		return nil
	}

	if status.Stale {
		_ = os.Remove(cfg.MonitorStatePath)
		_ = os.Remove(cfg.MonitorStopPath)
	}

	output.Println("Clipboard monitoring: stopped")
	return nil
}

// runMonitorCommand is the hidden worker entrypoint used by `ch start` to spawn
// the long-running clipboard monitoring child process.
func runMonitorCommand(logger commandLogger, cfg config.Config) error {
	logger.Info("running monitor command")
	return runMonitorWorker(logger, cfg)
}

// newMonitorController builds the controller used by start, stop, and status commands.
var newMonitorController = func(logger commandLogger, cfg config.Config) (monitorController, error) {
	controller, err := monitor.NewController(monitor.ControllerConfig{
		StatePath: cfg.MonitorStatePath,
		StopPath:  cfg.MonitorStopPath,
		Logger:    logger,
	})
	return controller, err
}

// monitorController is the lifecycle behavior required by monitor CLI commands.
type monitorController interface {
	// StartMonitor starts the worker and reports whether a new process was started.
	StartMonitor() (bool, error)
	// StopMonitor stops the worker and reports whether a running process was found.
	StopMonitor() (bool, error)
	// GetMonitorStatus returns the current worker status.
	GetMonitorStatus() (monitor.Status, error)
}

// runMonitorWorker runs the hidden monitor worker process until it is stopped or errors.
var runMonitorWorker = func(logger commandLogger, cfg config.Config) error {
	return monitor.RunWorker(logger, cfg.MonitorStatePath, cfg.MonitorStopPath)
}
