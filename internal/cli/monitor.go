package cli

import (
	"fmt"
	"os"

	"github.com/hoza-labs/clip-history/internal/config"
	"github.com/hoza-labs/clip-history/internal/logging"
	"github.com/hoza-labs/clip-history/internal/monitor"
)

var newMonitorService = func(cfg config.Config) monitor.Service {
	service := monitor.NewService(cfg.MonitorStatePath, cfg.MonitorStopPath)
	service.StartProcess = func() error {
		return startMonitorProcess()
	}
	service.KillProcess = killMonitorProcess
	return service
}

var runClipboardMonitor = func(logger logging.Logger, cfg config.Config) error {
	return monitor.Run(logger, cfg.MonitorStatePath, cfg.MonitorStopPath)
}

func runStartCommand(logger logging.Logger, cfg config.Config) error {
	logger.Info("running start command")

	started, err := newMonitorService(cfg).Start()
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

	stopped, err := newMonitorService(cfg).Stop()
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

	status, err := newMonitorService(cfg).Status()
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

func runMonitorCommand(logger logging.Logger, cfg config.Config) error {
	logger.Info("running monitor command")
	return runClipboardMonitor(logger, cfg)
}
