package cli

import (
	"fmt"
	"os"

	"github.com/hoza-labs/clip-history/internal/config"
	"github.com/hoza-labs/clip-history/internal/logging"
)

func Execute() error {
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

	if len(os.Args) < 2 {
		return printRootHelp(logger)
	}

	switch os.Args[1] {
	case "help", "--help", "-h":
		return printRootHelp(logger)
	case "config":
		return runConfigCommand(logger, cfg)
	case "status":
		return runStatusCommand(logger, cfg)
	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func printRootHelp(logger logging.Logger) error {
	logger.Info("displaying root help")
	fmt.Println("clip-history (ch)")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  ch <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  config      Show or initialize configuration")
	fmt.Println("  status      Show current clipboard history status")
	return nil
}
