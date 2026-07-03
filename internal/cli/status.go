package cli

import (
	"fmt"

	"github.com/hoza-labs/clip-history/internal/config"
	"github.com/hoza-labs/clip-history/internal/logging"
)

func runStatusCommand(logger logging.Logger, cfg config.Config) error {
	logger.Info("running status command")
	fmt.Printf("Clipboard history directory: %s\n", cfg.HistoryDir)
	return nil
}
