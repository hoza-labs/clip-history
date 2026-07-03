//go:build !windows

package cli

import "errors"

func startMonitorProcess() error {
	return errors.New("clipboard monitoring is only supported on Windows")
}

func killMonitorProcess(pid int) error {
	return nil
}
