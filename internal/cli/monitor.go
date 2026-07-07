package cli

import (
	"os"

	"github.com/hoza-labs/clip-history/internal/config"
	"github.com/hoza-labs/clip-history/internal/monitor"
)

type monitorController interface {
	StartMonitor() (bool, error)
	StopMonitor() (bool, error)
	GetMonitorStatus() (monitor.Status, error)
}

var newMonitorController = func(cfg config.Config) (monitorController, error) {
	controller, err := monitor.NewController(monitor.ControllerConfig{
		StatePath: cfg.MonitorStatePath,
		StopPath:  cfg.MonitorStopPath,
	})
	return controller, err
}

var runMonitorWorker = func(logger commandLogger, cfg config.Config) error {
	return monitor.RunWorker(logger, cfg.MonitorStatePath, cfg.MonitorStopPath)
}

func runStartCommand(logger commandLogger, output commandOutput, cfg config.Config) error {
	logger.Info("running start command")

	controller, err := newMonitorController(cfg)
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

func runStopCommand(logger commandLogger, output commandOutput, cfg config.Config) error {
	logger.Info("running stop command")

	controller, err := newMonitorController(cfg)
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

func runStatusCommand(logger commandLogger, output commandOutput, cfg config.Config) error {
	logger.Info("running status command")

	controller, err := newMonitorController(cfg)
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
