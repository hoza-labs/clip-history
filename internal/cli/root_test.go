package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestExecuteShowsHelpWithoutCommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ch"}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Execute()
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got := buf.String(); got == "" {
		t.Fatal("expected help output to be written")
	}
}

func TestExecuteShowsVersion(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	oldCurrentVersion := currentVersion
	currentVersion = func() string { return "test-version" }
	t.Cleanup(func() { currentVersion = oldCurrentVersion })

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ch", "version"}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Execute()
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if got := strings.TrimSpace(buf.String()); got != "test-version" {
		t.Fatalf("version output = %q, want %q", got, "test-version")
	}
}
