package logging

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestNewLoggerCreatesLogFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	freezeTime(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local))

	logger, err := NewLogger(dir, 14)
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	defer logger.Close()

	logPath := filepath.Join(dir, "clip-history-2026-06-07.log")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}

	logger.Info("hello")
}

func TestLoggerWritesTimestampPIDLevelAndMessage(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	freezeTime(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local))

	logger, err := NewLogger(dir, 14)
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	defer logger.Close()

	logger.Error("boom")

	data, err := os.ReadFile(filepath.Join(dir, "clip-history-2026-06-07.log"))
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
	freezeTime(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local))

	first, err := NewLogger(dir, 14)
	if err != nil {
		t.Fatalf("NewLogger first returned error: %v", err)
	}
	defer first.Close()

	second, err := NewLogger(dir, 14)
	if err != nil {
		t.Fatalf("NewLogger second returned error: %v", err)
	}
	defer second.Close()

	for i := 0; i < 25; i++ {
		first.Info("first")
		second.Error("second")
	}

	data, err := os.ReadFile(filepath.Join(dir, "clip-history-2026-06-07.log"))
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

func TestNewLoggerRemovesOnlyMatchingLogsOlderThanRetention(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create log dir: %v", err)
	}
	freezeTime(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local))

	files := []string{
		"clip-history-2026-20-06.log",
		"clip-history-2026-22-06.log",
		"clip-history-2026-06-07.log",
		"clip-history.log",
		"clip-history-2026-20-06.txt",
		"other-2026-20-06.log",
	}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	logger, err := NewLogger(dir, 14)
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	defer logger.Close()

	assertMissing(t, filepath.Join(dir, "clip-history-2026-20-06.log"))
	for _, name := range files[1:] {
		assertExists(t, filepath.Join(dir, name))
	}
}

func TestNewLoggerRejectsNegativeRetention(t *testing.T) {
	_, err := NewLogger(filepath.Join(t.TempDir(), "logs"), -1)
	if err == nil {
		t.Fatal("expected NewLogger to reject negative retention")
	}
}

func freezeTime(t *testing.T, value time.Time) {
	t.Helper()
	oldNow := now
	now = func() time.Time { return value }
	t.Cleanup(func() { now = oldNow })
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", filepath.Base(path), err)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be removed, stat err = %v", filepath.Base(path), err)
	}
}
