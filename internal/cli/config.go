package cli

import (
	"fmt"

	"github.com/hoza-labs/clip-history/internal/config"
)

func runConfigCommand(logger commandLogger, cfg config.Config) error {
	logger.Info("running config command")
	fmt.Printf("Config path: %s\n", cfg.Path)
	fmt.Printf("Log dir: %s\n", cfg.LogDir)
	fmt.Printf("Log timezone: %s\n", cfg.LogTimezone)
	return nil
}
