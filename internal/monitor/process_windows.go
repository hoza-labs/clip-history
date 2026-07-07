//go:build windows

package monitor

import (
	"os"
	"os/exec"
	"syscall"
)

// createNoWindow hides the spawned worker console window on Windows.
const createNoWindow = 0x08000000

// startMonitorProcess starts a hidden child process running the private monitor command.
var startMonitorProcess = func() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(exe, "monitor")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: createNoWindow,
	}

	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer devNull.Close()

	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Process.Release()
}

// killMonitorProcess kills pid as a last resort when the worker ignores the stop signal.
var killMonitorProcess = func(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Kill()
}
