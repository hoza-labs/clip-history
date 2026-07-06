package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
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

func TestExecuteLogsCommandErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ch", "bogus"}

	err := Execute()
	if err == nil {
		t.Fatal("expected Execute to return an error")
	}

	logs, globErr := filepath.Glob(filepath.Join(home, ".clip-history", "logs", "clip-history-*.log"))
	if globErr != nil {
		t.Fatalf("glob log file: %v", globErr)
	}
	if len(logs) != 1 {
		t.Fatalf("log file count = %d, want 1", len(logs))
	}

	data, readErr := os.ReadFile(logs[0])
	if readErr != nil {
		t.Fatalf("read log file: %v", readErr)
	}

	if !strings.Contains(string(data), "[ERROR] unknown command: bogus") {
		t.Fatalf("log file = %q, want command error", string(data))
	}
}
