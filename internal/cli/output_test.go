package cli

import (
	"bytes"
	"reflect"
	"testing"
)

// recordingLogger captures logged output for tests.
type recordingLogger struct {
	infos []string
}

// Info records message.
func (r *recordingLogger) Info(message string) {
	r.infos = append(r.infos, message)
}

// Error ignores error messages.
func (r *recordingLogger) Error(string) {}

// Debug ignores debug messages.
func (r *recordingLogger) Debug(string) {}

// TestLoggedOutputWritesToStdoutAndLogger verifies output is mirrored to the log.
func TestLoggedOutputWritesToStdoutAndLogger(t *testing.T) {
	var stdout bytes.Buffer
	logger := &recordingLogger{}
	output := loggedOutput{
		writer: &stdout,
		logger: logger,
	}

	output.Printf("Hello %s\n", "there")
	output.Println("Done")

	if got, want := stdout.String(), "Hello there\nDone\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}

	wantLogs := []string{"> Hello there", "> Done"}
	if !reflect.DeepEqual(logger.infos, wantLogs) {
		t.Fatalf("logger infos = %#v, want %#v", logger.infos, wantLogs)
	}
}
