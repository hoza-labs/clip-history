//go:build !windows

package monitor

import (
	"errors"

	"github.com/hoza-labs/clip-history/internal/logging"
)

func RunWorker(logger logging.Logger, statePath, stopPath string) error {
	return errors.New("cannot run monitor worker: clipboard monitoring is only supported on Windows")
}
