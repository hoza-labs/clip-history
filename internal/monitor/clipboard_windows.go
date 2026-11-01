//go:build windows

package monitor

import (
	"errors"
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

// clipboardReader inspects format metadata while the caller owns the open clipboard.
// All calls must run serially on the window's OS thread; Close follows a successful Open.
type clipboardReader interface {
	Open() error
	Close() error
	Next(uint32) (uint32, error)
	Name(uint32) string
}

// logClipboardFormats records available format IDs and names without reading clip contents.
func logClipboardFormats(logger workerLogger, clipboard clipboardReader) {
	attempts, err := openClipboardWithRetry(clipboard.Open, time.Sleep)
	if err != nil {
		logger.Info(fmt.Sprintf("clipboard format enumeration failed: %v", err))
		return
	}
	defer func() {
		if err := clipboard.Close(); err != nil {
			logger.Info(fmt.Sprintf("clipboard close failed: %v", err))
		}
	}()
	if attempts > 1 {
		logger.Info(fmt.Sprintf("clipboard access acquired after %d attempts", attempts))
	}

	count := 0
	for format := uint32(0); ; {
		next, err := clipboard.Next(format)
		if err != nil {
			logger.Info(fmt.Sprintf("clipboard format enumeration failed after %d format(s): %v", count, err))
			return
		}
		if next == 0 {
			logger.Info(fmt.Sprintf("clipboard format enumeration complete: %d format(s)", count))
			return
		}
		logger.Info(fmt.Sprintf("clipboard format: id=%d name=%q", next, clipboard.Name(next)))
		count++
		format = next
	}
}

const (
	// clipboardOpenAttempts bounds retries so contention cannot indefinitely block the message loop.
	clipboardOpenAttempts = 10
	// clipboardOpenRetryDelay gives other applications time to release the clipboard between attempts.
	clipboardOpenRetryDelay = 25 * time.Millisecond
)

// openClipboardWithRetry retries access-denied errors on the calling OS thread.
// It returns the number of attempts and the final error; sleep controls the delay for tests.
func openClipboardWithRetry(open func() error, sleep func(time.Duration)) (int, error) {
	for attempt := 1; ; attempt++ {
		err := open()
		if err == nil {
			return attempt, nil
		}
		if !errors.Is(err, syscall.ERROR_ACCESS_DENIED) {
			return attempt, err
		}
		if attempt == clipboardOpenAttempts {
			return attempt, fmt.Errorf("clipboard access unavailable after %d attempts: %w", attempt, err)
		}
		sleep(clipboardOpenRetryDelay)
	}
}

// systemClipboard reads format metadata using the worker window's handle.
type systemClipboard struct {
	hwnd uintptr
}

// Open acquires clipboard access or returns an error if another application holds it.
func (c systemClipboard) Open() error {
	if ok, _, err := procOpenClipboard.Call(c.hwnd); ok == 0 {
		return fmt.Errorf("open clipboard: %w", err)
	}
	return nil
}

// Close releases clipboard access.
func (c systemClipboard) Close() error {
	if ok, _, err := procCloseClipboard.Call(); ok == 0 {
		return fmt.Errorf("close clipboard: %w", err)
	}
	return nil
}

// Next returns the format after previous, or zero when enumeration completes.
func (c systemClipboard) Next(previous uint32) (uint32, error) {
	format, _, err := procEnumClipboardFormats.Call(uintptr(previous))
	if format == 0 && err != syscall.Errno(0) {
		return 0, fmt.Errorf("enumerate clipboard formats: %w", err)
	}
	return uint32(format), nil
}

// Name returns a standard or registered format name, with a fallback for unnamed formats.
func (c systemClipboard) Name(format uint32) string {
	if name, ok := standardClipboardFormats[format]; ok {
		return name
	}
	// Registered clipboard format names are limited to 255 characters.
	var buffer [256]uint16
	length, _, _ := procGetClipboardFormatNameW.Call(uintptr(format), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	if length != 0 {
		return syscall.UTF16ToString(buffer[:length])
	}
	return "unknown"
}

// standardClipboardFormats names predefined formats, which GetClipboardFormatNameW does not name.
var standardClipboardFormats = map[uint32]string{
	1: "CF_TEXT", 2: "CF_BITMAP", 3: "CF_METAFILEPICT", 4: "CF_SYLK",
	5: "CF_DIF", 6: "CF_TIFF", 7: "CF_OEMTEXT", 8: "CF_DIB",
	9: "CF_PALETTE", 10: "CF_PENDATA", 11: "CF_RIFF", 12: "CF_WAVE",
	13: "CF_UNICODETEXT", 14: "CF_ENHMETAFILE", 15: "CF_HDROP",
	16: "CF_LOCALE", 17: "CF_DIBV5", 0x80: "CF_OWNERDISPLAY",
	0x81: "CF_DSPTEXT", 0x82: "CF_DSPBITMAP", 0x83: "CF_DSPMETAFILEPICT",
	0x8e: "CF_DSPENHMETAFILE",
}

var (
	// procOpenClipboard acquires access to enumerate clipboard formats.
	procOpenClipboard = user32.NewProc("OpenClipboard")
	// procCloseClipboard releases clipboard access.
	procCloseClipboard = user32.NewProc("CloseClipboard")
	// procEnumClipboardFormats retrieves successive available format IDs.
	procEnumClipboardFormats = user32.NewProc("EnumClipboardFormats")
	// procGetClipboardFormatNameW retrieves a registered format's UTF-16 name.
	procGetClipboardFormatNameW = user32.NewProc("GetClipboardFormatNameW")
)
