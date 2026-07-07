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

	logger, err := NewLogger(dir, 14, "local")
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	defer logger.Close()

	logPath := filepath.Join(dir, logFileName(time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local)))
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}

	logger.Info("hello")
}

func TestLoggerWritesTimestampPIDLevelAndMessage(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	freezeTime(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local))

	logger, err := NewLogger(dir, 14, "local")
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	defer logger.Close()

	logger.Error("boom")

	data, err := os.ReadFile(filepath.Join(dir, logFileName(time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local))))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	line := strings.TrimSpace(string(data))
	pattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(Z|[+-]\d{2}:\d{2}) \[\s{0,6}\d{1,}\] \[ERROR\] boom$`
	if !regexp.MustCompile(pattern).MatchString(line) {
		t.Fatalf("log line = %q, want timestamp [padded pid] level message", line)
	}
}

func TestLoggerWritesCompleteLinesAcrossMultipleLoggerInstances(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	freezeTime(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local))

	first, err := NewLogger(dir, 14, "local")
	if err != nil {
		t.Fatalf("NewLogger first returned error: %v", err)
	}
	defer first.Close()

	second, err := NewLogger(dir, 14, "local")
	if err != nil {
		t.Fatalf("NewLogger second returned error: %v", err)
	}
	defer second.Close()

	for i := 0; i < 25; i++ {
		first.Info("first")
		second.Error("second")
	}

	data, err := os.ReadFile(filepath.Join(dir, logFileName(time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local))))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 50 {
		t.Fatalf("line count = %d, want 50", len(lines))
	}

	pattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(Z|[+-]\d{2}:\d{2}) \[\s{0,6}\d{1,}\] \[(INFO |ERROR)\] (first|second)$`)
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
		logFileName(time.Date(2026, 6, 20, 12, 0, 0, 0, time.Local)),
		logFileName(time.Date(2026, 6, 22, 12, 0, 0, 0, time.Local)),
		logFileName(time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local)),
		"clip-history.log",
		strings.TrimSuffix(logFileName(time.Date(2026, 6, 20, 12, 0, 0, 0, time.Local)), ".log") + ".txt",
		"other-2026-06-20.log",
	}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	logger, err := NewLogger(dir, 14, "local")
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	defer logger.Close()

	assertMissing(t, filepath.Join(dir, logFileName(time.Date(2026, 6, 20, 12, 0, 0, 0, time.Local))))
	for _, name := range files[1:] {
		assertExists(t, filepath.Join(dir, name))
	}
}

func TestNewLoggerRejectsNegativeRetention(t *testing.T) {
	_, err := NewLogger(filepath.Join(t.TempDir(), "logs"), -1, "local")
	if err == nil {
		t.Fatal("expected NewLogger to reject negative retention")
	}
}

func TestLoggerUsesConfiguredTimezoneForFileAndEntries(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	freezeTime(t, time.Date(2026, 7, 6, 6, 30, 0, 0, time.UTC))

	logger, err := NewLogger(dir, 14, "America/Los_Angeles")
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	defer logger.Close()

	logger.Info("hello")

	data, err := os.ReadFile(filepath.Join(dir, logFileName(time.Date(2026, 7, 5, 23, 30, 0, 0, time.FixedZone("PDT", -7*60*60)))))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.HasPrefix(string(data), "2026-07-05T23:30:00-07:00") {
		t.Fatalf("log line = %q, want configured timezone timestamp", string(data))
	}
}

func TestNewLoggerRejectsInvalidTimezone(t *testing.T) {
	_, err := NewLogger(filepath.Join(t.TempDir(), "logs"), 14, "Not/AZone")
	if err == nil {
		t.Fatal("expected NewLogger to reject invalid timezone")
	}
}

func TestLoggerRotatesFilesAndRunsRetentionWhenDateChanges(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create log dir: %v", err)
	}

	currentTime := time.Date(2026, 7, 6, 23, 59, 59, 0, time.Local)
	oldNow := now
	now = func() time.Time { return currentTime }
	t.Cleanup(func() { now = oldNow })

	expiredLog := logFileName(time.Date(2026, 7, 5, 12, 0, 0, 0, time.Local))
	if err := os.WriteFile(filepath.Join(dir, expiredLog), []byte("old"), 0o644); err != nil {
		t.Fatalf("write expired log: %v", err)
	}

	logger, err := NewLogger(dir, 1, "local")
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	defer logger.Close()

	logger.Info("before midnight")
	currentTime = time.Date(2026, 7, 7, 0, 0, 1, 0, time.Local)
	logger.Info("after midnight")

	beforeData, err := os.ReadFile(filepath.Join(dir, logFileName(time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local))))
	if err != nil {
		t.Fatalf("read pre-rotation log: %v", err)
	}
	if !strings.Contains(string(beforeData), "before midnight") {
		t.Fatalf("pre-rotation log = %q, want before midnight entry", string(beforeData))
	}

	afterData, err := os.ReadFile(filepath.Join(dir, logFileName(time.Date(2026, 7, 7, 12, 0, 0, 0, time.Local))))
	if err != nil {
		t.Fatalf("read rotated log: %v", err)
	}
	if !strings.Contains(string(afterData), "after midnight") {
		t.Fatalf("rotated log = %q, want after midnight entry", string(afterData))
	}
	assertMissing(t, filepath.Join(dir, expiredLog))
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

func logFileName(t time.Time) string {
	return "clip-history-" + t.Format(logFileDateLayout) + ".log"
}
