//go:build !windows

package monitor

import "errors"

var startMonitorProcess = func() error {
	return errors.New("cannot start monitor worker: clipboard monitoring is only supported on Windows")
}

var killMonitorProcess = func(pid int) error {
	return nil
}
