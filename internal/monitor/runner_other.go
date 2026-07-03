//go:build !windows

package monitor

import (
	"errors"

	"github.com/hoza-labs/clip-history/internal/logging"
)

func Run(logger logging.Logger, statePath, stopPath string) error {
	return errors.New("clipboard monitoring is only supported on Windows")
}
