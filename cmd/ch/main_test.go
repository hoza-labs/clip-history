package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestRunDoesNothingWhenExecuteSucceeds(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := -1

	run(
		func() error { return nil },
		&stderr,
		func(code int) { exitCode = code },
	)

	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}

	if exitCode != -1 {
		t.Fatalf("exit code = %d, want exit not called", exitCode)
	}
}

func TestRunPrintsErrorAndExitsNonZeroWhenExecuteFails(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := -1

	run(
		func() error { return errors.New("boom") },
		&stderr,
		func(code int) { exitCode = code },
	)

	if got := stderr.String(); got != "boom\n" {
		t.Fatalf("stderr = %q, want %q", got, "boom\n")
	}

	if exitCode != 1 {
		t.Fatalf("exit code = %d, want %d", exitCode, 1)
	}
}
