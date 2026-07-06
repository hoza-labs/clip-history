//go:build !windows

package monitor

import (
	"errors"
)

func RunWorker(logger workerLogger, statePath, stopPath string) error {
	return errors.New("cannot run monitor worker: clipboard monitoring is only supported on Windows")
}
