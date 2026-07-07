package cli

import (
	"bytes"
	"reflect"
	"testing"
)

type recordingLogger struct {
	infos []string
}

func (r *recordingLogger) Info(message string) {
	r.infos = append(r.infos, message)
}

func (r *recordingLogger) Error(string) {}
func (r *recordingLogger) Debug(string) {}

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

	wantLogs := []string{"Hello there", "Done"}
	if !reflect.DeepEqual(logger.infos, wantLogs) {
		t.Fatalf("logger infos = %#v, want %#v", logger.infos, wantLogs)
	}
}
