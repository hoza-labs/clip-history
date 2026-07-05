//go:build windows

package monitor

import (
	"os"
	"os/exec"
	"syscall"
)

const createNoWindow = 0x08000000

// startMonitorProcess starts a new instance of the program in the background.
// The new instance will run the clipboard monitor until it receives a "stop" signal.
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

// killMonitorProcess sends a "kill" signal to the process with the given PID. The Controller uses this
// method as a last resort to stop the clipboard monitor if it does not respond to the "stop" signal.
var killMonitorProcess = func(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Kill()
}
