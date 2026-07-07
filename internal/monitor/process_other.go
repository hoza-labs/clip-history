//go:build !windows

package monitor

import "errors"

// startMonitorProcess returns an unsupported-platform error outside Windows.
var startMonitorProcess = func() error {
	return errors.New("cannot start monitor worker: clipboard monitoring is only supported on Windows")
}

// killMonitorProcess is a no-op outside Windows because no worker process can be started.
var killMonitorProcess = func(pid int) error {
	return nil
}
