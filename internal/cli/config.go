package cli

import (
	"fmt"

	"github.com/hoza-labs/clip-history/internal/config"
	"github.com/pelletier/go-toml/v2"
)

// runConfigCommand prints every effective setting as TOML and returns encoding errors.
func runConfigCommand(logger commandLogger, output commandOutput, cfg config.Config) error {
	logger.Info("running config command")
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode effective config: %w", err)
	}
	output.Printf("%s", data)
	return nil
}
