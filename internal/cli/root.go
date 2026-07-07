// Package cli implements the ch command tree and command-specific protocols.
//
// Protocol: Execute handles commands from os.Args; most commands load config and
// create a logger that must be closed, while version/help return before logging.
// The private "monitor" command is a hidden worker entrypoint spawned by "start"
// and should not be exposed in help.
package cli

import (
	"fmt"
	"os"

	"github.com/hoza-labs/clip-history/internal/config"
	"github.com/hoza-labs/clip-history/internal/logging"
)

// commandLogger is the logging behavior required by command handlers.
type commandLogger interface {
	// Info records an informational message.
	Info(string)
	// Error records an error message.
	Error(string)
	// Debug records a debug message.
	Debug(string)
}

// Execute runs the command named in os.Args and returns any command or setup error.
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

	err = runCommand(os.Args[1], logger, newLoggedOutput(logger), cfg)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	logger.Info("clip-history exit")
	return nil
}

// runCommand dispatches command using logger, output, and cfg, returning an unknown-command error when needed.
func runCommand(command string, logger commandLogger, output commandOutput, cfg config.Config) error {
	switch command {
	case "config":
		return runConfigCommand(logger, output, cfg)
	case "start":
		// this is the public command that causes the Controller to spawn a separate
		// instance of the program to run in the background until told to stop.
		return runStartCommand(logger, output, cfg)
	case "stop":
		// this is the public command that causes the Controller to send a "stop" signal
		// to the clipboard monitor (the separate instance running in the background).
		return runStopCommand(logger, output, cfg)
	case "status":
		// this is the public command that causes the Controller to obtain the current status
		// of the clipboard monitor (the separate instance running in the background).
		return runStatusCommand(logger, output, cfg)
	case "monitor":
		// This is a "hidden" command that actually runs the clipboard monitor in the background.
		// It does not return until it receives a "stop" external signal.
		return runMonitorCommand(logger, cfg)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// printRootHelp writes the public command summary and returns nil.
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
