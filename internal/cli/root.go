package cli

import (
	"fmt"
	"os"

	"github.com/hoza-labs/clip-history/internal/config"
	"github.com/hoza-labs/clip-history/internal/logging"
)

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

	logger, err := logging.NewLogger(cfg.LogDir)
	if err != nil {
		return err
	}
	defer logger.Close()

	logger.Info("clip-history started")

	switch os.Args[1] {
	case "config":
		return runConfigCommand(logger, cfg)
	case "start":
		return runStartCommand(logger, cfg)
	case "stop":
		return runStopCommand(logger, cfg)
	case "status":
		return runStatusCommand(logger, cfg)
	case "monitor":
		// this is an internal command, used to spawn the background monitoring process
		return runMonitorCommand(logger, cfg)
	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
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
