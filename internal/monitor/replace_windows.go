//go:build windows

package monitor

import (
	"errors"
	"os"
	"syscall"
	"time"
	"unsafe"
)

const (
	// moveFileReplaceExisting allows MoveFileExW to replace an existing destination.
	moveFileReplaceExisting = 0x1
	// moveFileWriteThrough waits for the replacement to be flushed to disk.
	moveFileWriteThrough = 0x8
	// replaceRetryCount bounds retries while another process briefly has the state file open.
	replaceRetryCount = 50
	// replaceRetryDelay is the pause between transient Windows sharing failures.
	replaceRetryDelay = 2 * time.Millisecond
	// errorSharingViolation reports that another process has the destination open without delete sharing.
	errorSharingViolation syscall.Errno = 32
)

var (
	// kernel32File loads the file-replacement procedure from kernel32.dll.
	kernel32File = syscall.NewLazyDLL("kernel32.dll")
	// procMoveFileExW atomically moves a file over an existing destination.
	procMoveFileExW = kernel32File.NewProc("MoveFileExW")
)

// readStateFile reads path, retrying transient sharing failures during atomic replacement.
func readStateFile(path string) ([]byte, error) {
	var lastErr error
	for attempt := range replaceRetryCount {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
		lastErr = err
		if !errors.Is(err, syscall.ERROR_ACCESS_DENIED) && !errors.Is(err, errorSharingViolation) {
			return nil, err
		}
		if attempt+1 < replaceRetryCount {
			time.Sleep(replaceRetryDelay)
		}
	}

	return nil, lastErr
}

// replaceFile atomically moves source over destination and flushes the operation to disk.
func replaceFile(source, destination string) error {
	sourcePtr, err := syscall.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	destinationPtr, err := syscall.UTF16PtrFromString(destination)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := range replaceRetryCount {
		ok, _, callErr := procMoveFileExW.Call(
			uintptr(unsafe.Pointer(sourcePtr)),
			uintptr(unsafe.Pointer(destinationPtr)),
			moveFileReplaceExisting|moveFileWriteThrough,
		)
		if ok != 0 {
			return nil
		}
		lastErr = callErr
		if !errors.Is(callErr, syscall.ERROR_ACCESS_DENIED) && !errors.Is(callErr, errorSharingViolation) {
			return callErr
		}
		if attempt+1 < replaceRetryCount {
			time.Sleep(replaceRetryDelay)
		}
	}

	return lastErr
}
