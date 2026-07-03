package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewLoggerCreatesLogFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")

	logger, err := NewLogger(dir)
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	defer logger.Close()

	logPath := filepath.Join(dir, "clip-history.log")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}

	logger.Info("hello")
}
