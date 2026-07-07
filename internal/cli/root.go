package cli

import (
	"fmt"
	"os"

	"github.com/hoza-labs/clip-history/internal/config"
	"github.com/hoza-labs/clip-history/internal/logging"
)

type commandLogger interface {
	Info(string)
	Error(string)
	Debug(string)
}

func Execute() error {
	if len(os.Args) < 2 {
		return printRootHelp()
	}

	switch os.Args[1] {
	case "help", "--help", "-h":
		return printRootHelp()
	case "version":
		return runVersionCommand()
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger, err := logging.NewLogger(cfg.LogDir, cfg.LogRetentionDays, cfg.LogTimezone, os.Args[1])
	if err != nil {
		return err
	}
	defer logger.Close()

	logger.Info("clip-history entry")

	err = runCommand(os.Args[1], logger, cfg)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	logger.Info("clip-history exit")
	return nil
}

func runCommand(command string, logger commandLogger, cfg config.Config) error {
	switch command {
	case "config":
		return runConfigCommand(logger, cfg)
	case "start":
		// this is the public command that causes the Controller to spawn a separate
		// instance of the program to run in the background until told to stop.
		return runStartCommand(logger, cfg)
	case "stop":
		// this is the public command that causes the Controller to send a "stop" signal
		// to the clipboard monitor (the separate instance running in the background).
		return runStopCommand(logger, cfg)
	case "status":
		// this is the public command that causes the Controller to obtain the current status
		// of the clipboard monitor (the separate instance running in the background).
		return runStatusCommand(logger, cfg)
	case "monitor":
		// This is a "hidden" command that actually runs the clipboard monitor in the background.
		// It does not return until it receives a "stop" external signal.
		return runMonitorCommand(logger, cfg)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func printRootHelp() error {
	fmt.Println("clip-history (ch)")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  ch <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  config      Show or initialize configuration")
	fmt.Println("  start       Start clipboard monitoring in the background")
	fmt.Println("  stop        Stop clipboard monitoring")
	fmt.Println("  status      Show whether clipboard monitoring is running")
	fmt.Println("  version     Show the current target release version")
	return nil
}
