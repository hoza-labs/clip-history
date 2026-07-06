package logging

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

func TestLoggerWritesTimestampPIDLevelAndMessage(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")

	logger, err := NewLogger(dir)
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	defer logger.Close()

	logger.Error("boom")

	data, err := os.ReadFile(filepath.Join(dir, "clip-history.log"))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	line := strings.TrimSpace(string(data))
	pattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z \[\s{0,6}\d{1,}\] \[ERROR\] boom$`
	if !regexp.MustCompile(pattern).MatchString(line) {
		t.Fatalf("log line = %q, want timestamp [padded pid] level message", line)
	}
}

func TestLoggerWritesCompleteLinesAcrossMultipleLoggerInstances(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")

	first, err := NewLogger(dir)
	if err != nil {
		t.Fatalf("NewLogger first returned error: %v", err)
	}
	defer first.Close()

	second, err := NewLogger(dir)
	if err != nil {
		t.Fatalf("NewLogger second returned error: %v", err)
	}
	defer second.Close()

	for i := 0; i < 25; i++ {
		first.Info("first")
		second.Error("second")
	}

	data, err := os.ReadFile(filepath.Join(dir, "clip-history.log"))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 50 {
		t.Fatalf("line count = %d, want 50", len(lines))
	}

	pattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z \[\s{0,6}\d{1,}\] \[(INFO |ERROR)\] (first|second)$`)
	for _, line := range lines {
		if !pattern.MatchString(line) {
			t.Fatalf("malformed log line = %q", line)
		}
	}
}
