package cli

import (
	"bytes"
	"io"
	"os"
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
