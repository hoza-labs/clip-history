package cli

import "github.com/hoza-labs/clip-history/internal/config"

func runConfigCommand(logger commandLogger, output commandOutput, cfg config.Config) error {
	logger.Info("running config command")
	output.Printf("Config path: %s\n", cfg.Path)
	output.Printf("Log dir: %s\n", cfg.LogDir)
	output.Printf("Log timezone: %s\n", cfg.LogTimezone)
	return nil
}
