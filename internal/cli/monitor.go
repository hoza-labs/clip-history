package cli

import (
	"fmt"
	"os"

	"github.com/hoza-labs/clip-history/internal/config"
	"github.com/hoza-labs/clip-history/internal/logging"
	"github.com/hoza-labs/clip-history/internal/monitor"
)

var newMonitorController = func(cfg config.Config) (monitor.Controller, error) {
	controller, err := monitor.NewController(monitor.ControllerConfig{
		StatePath:    cfg.MonitorStatePath,
		StopPath:     cfg.MonitorStopPath,
		StartProcess: startMonitorProcess,
		KillProcess:  killMonitorProcess,
	})
	return controller, err
}

var runMonitorWorker = func(logger logging.Logger, cfg config.Config) error {
	return monitor.RunWorker(logger, cfg.MonitorStatePath, cfg.MonitorStopPath)
}

func runStartCommand(logger logging.Logger, cfg config.Config) error {
	logger.Info("running start command")

	controller, err := newMonitorController(cfg)
	if err != nil {
		return err
	}
	started, err := controller.Start()
	if err != nil {
		return err
	}
	if !started {
		fmt.Println("Clipboard monitoring is already running.")
		return nil
	}

	fmt.Println("Clipboard monitoring started.")
	return nil
}

func runStopCommand(logger logging.Logger, cfg config.Config) error {
	logger.Info("running stop command")

	controller, err := newMonitorController(cfg)
	if err != nil {
		return err
	}
	stopped, err := controller.Stop()
	if err != nil {
		return err
	}
	if !stopped {
		fmt.Println("Clipboard monitoring is not running.")
		return nil
	}

	fmt.Println("Clipboard monitoring stopped.")
	return nil
}

func runStatusCommand(logger logging.Logger, cfg config.Config) error {
	logger.Info("running status command")

	controller, err := newMonitorController(cfg)
	if err != nil {
		return err
	}
	status, err := controller.Status()
	if err != nil {
		return err
	}

	fmt.Printf("Clipboard history directory: %s\n", cfg.HistoryDir)
	if status.Running {
		fmt.Printf("Clipboard monitoring: running (pid %d)\n", status.PID)
		return nil
	}

	if status.Stale {
		_ = os.Remove(cfg.MonitorStatePath)
		_ = os.Remove(cfg.MonitorStopPath)
	}

	fmt.Println("Clipboard monitoring: stopped")
	return nil
}

// runMonitorCommand is the hidden worker entrypoint used by `ch start` to spawn
// the long-running clipboard monitoring child process.
func runMonitorCommand(logger logging.Logger, cfg config.Config) error {
	logger.Info("running monitor command")
	return runMonitorWorker(logger, cfg)
}
