//go:build windows

package monitor

import (
	"errors"
	"fmt"
	"reflect"
	"syscall"
	"testing"
	"time"
)

// TestOpenClipboardWithRetry verifies recovery, bounded waiting, and immediate non-contention failures.
func TestOpenClipboardWithRetry(t *testing.T) {
	for _, tc := range []struct {
		name       string
		failures   int
		err        error
		wantCalls  int
		wantSleeps int
		wantErr    bool
	}{
		{name: "given immediate access, does not wait", wantCalls: 1},
		{name: "given transient contention, retries until access succeeds", failures: 2,
			err: fmt.Errorf("open clipboard: %w", syscall.ERROR_ACCESS_DENIED), wantCalls: 3, wantSleeps: 2},
		{name: "given persistent contention, stops after the retry limit", failures: 100,
			err: syscall.ERROR_ACCESS_DENIED, wantCalls: clipboardOpenAttempts, wantSleeps: clipboardOpenAttempts - 1, wantErr: true},
		{name: "given another error, fails immediately", failures: 1,
			err: syscall.Errno(6), wantCalls: 1, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls, sleeps := 0, 0
			attempts, err := openClipboardWithRetry(func() error {
				calls++
				if calls <= tc.failures {
					return tc.err
				}
				return nil
			}, func(delay time.Duration) {
				sleeps++
				if delay != clipboardOpenRetryDelay {
					t.Fatalf("retry delay = %v, want %v", delay, clipboardOpenRetryDelay)
				}
			})
			if calls != tc.wantCalls || attempts != tc.wantCalls || sleeps != tc.wantSleeps {
				t.Fatalf("calls=%d attempts=%d sleeps=%d; want attempts=%d sleeps=%d", calls, attempts, sleeps, tc.wantCalls, tc.wantSleeps)
			}
			if (err != nil) != tc.wantErr || (tc.wantErr && !errors.Is(err, tc.err)) {
				t.Fatalf("error = %v, want failure=%v preserving %v", err, tc.wantErr, tc.err)
			}
		})
	}
}

// TestLogClipboardFormats verifies metadata logging and clipboard release without touching the desktop clipboard.
func TestLogClipboardFormats(t *testing.T) {
	for _, tc := range []struct {
		name      string
		formats   []uint32
		openErr   error
		nextErr   error
		closeErr  error
		want      []string
		wantClose bool
	}{
		{
			name:    "given standard and registered formats, logs each ID and quoted name",
			formats: []uint32{13, 49152}, wantClose: true,
			want: []string{
				`clipboard format: id=13 name="CF_UNICODETEXT"`,
				`clipboard format: id=49152 name="HTML Format"`,
				"clipboard format enumeration complete: 2 format(s)",
			},
		},
		{
			name: "given an empty clipboard, logs zero formats", wantClose: true,
			want: []string{"clipboard format enumeration complete: 0 format(s)"},
		},
		{
			name:    "given a busy clipboard, logs the failure without closing it",
			openErr: errors.New("busy"),
			want:    []string{"clipboard format enumeration failed: busy"},
		},
		{
			name:    "given an enumeration failure, preserves earlier metadata and closes",
			formats: []uint32{13}, nextErr: errors.New("unavailable"), wantClose: true,
			want: []string{
				`clipboard format: id=13 name="CF_UNICODETEXT"`,
				"clipboard format enumeration failed after 1 format(s): unavailable",
			},
		},
		{
			name: "given a close failure, logs it", closeErr: errors.New("close error"), wantClose: true,
			want: []string{"clipboard format enumeration complete: 0 format(s)", "clipboard close failed: close error"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clipboard := &fakeClipboard{formats: tc.formats, openErr: tc.openErr, nextErr: tc.nextErr, closeErr: tc.closeErr}
			logger := &formatLogger{}
			logClipboardFormats(logger, clipboard)
			if !reflect.DeepEqual(logger.messages, tc.want) {
				t.Fatalf("messages = %q, want %q", logger.messages, tc.want)
			}
			if clipboard.closed != tc.wantClose {
				t.Fatalf("clipboard closed = %v, want %v", clipboard.closed, tc.wantClose)
			}
		})
	}
}

// TestSystemClipboardName verifies predefined and unnamed format handling without clipboard access.
func TestSystemClipboardName(t *testing.T) {
	clipboard := systemClipboard{}
	for format, want := range map[uint32]string{13: "CF_UNICODETEXT", 15: "CF_HDROP", 0: "unknown"} {
		if got := clipboard.Name(format); got != want {
			t.Errorf("Name(%d) = %q, want %q", format, got, want)
		}
	}
}

// fakeClipboard supplies metadata and failures without accessing the real clipboard.
type fakeClipboard struct {
	formats  []uint32
	openErr  error
	nextErr  error
	closeErr error
	closed   bool
}

// Open returns the configured acquisition error.
func (c *fakeClipboard) Open() error { return c.openErr }

// Close records release and returns the configured release error.
func (c *fakeClipboard) Close() error { c.closed = true; return c.closeErr }

// Next returns the next configured format or the terminal error.
func (c *fakeClipboard) Next(previous uint32) (uint32, error) {
	for i, format := range c.formats {
		if previous == 0 || (i > 0 && c.formats[i-1] == previous) {
			return format, nil
		}
	}
	return 0, c.nextErr
}

// Name supplies names for the test formats.
func (c *fakeClipboard) Name(format uint32) string {
	if format == 49152 {
		return "HTML Format"
	}
	return systemClipboard{}.Name(format)
}

// formatLogger captures clipboard metadata messages for assertions.
type formatLogger struct{ messages []string }

// Info appends one captured message.
func (l *formatLogger) Info(message string) { l.messages = append(l.messages, message) }
