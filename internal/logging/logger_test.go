package logging

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestNewLoggerCreatesLogFile verifies construction opens the current daily log.
func TestNewLoggerCreatesLogFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	freezeTime(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local))

	logger, err := NewLogger(dir, 14, "local", "test")
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

// TestLoggerWritesTimestampPIDLevelAndMessage verifies log entry formatting.
func TestLoggerWritesTimestampPIDLevelAndMessage(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	freezeTime(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local))

	logger, err := NewLogger(dir, 14, "local", "config")
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	defer logger.Close()

	logger.Error("boom")

	data, err := os.ReadFile(filepath.Join(dir, logFileName(time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local))))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	line := lines[len(lines)-1]
	pattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}(Z|[+-]\d{2}:\d{2}) \[\s{0,6}\d{1,}\] \[config\s{4}\] \[ERROR\] boom$`
	if !regexp.MustCompile(pattern).MatchString(line) {
		t.Fatalf("log line = %q, want timestamp [padded pid] level message", line)
	}
}

// TestLoggerWritesCompleteLinesAcrossMultipleLoggerInstances verifies concurrent files stay line-oriented.
func TestLoggerWritesCompleteLinesAcrossMultipleLoggerInstances(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	freezeTime(t, time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local))

	first, err := NewLogger(dir, 14, "local", "first")
	if err != nil {
		t.Fatalf("NewLogger first returned error: %v", err)
	}
	defer first.Close()

	second, err := NewLogger(dir, 14, "local", "second")
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
	if len(lines) != 54 {
		t.Fatalf("line count = %d, want 54", len(lines))
	}

	pattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}(Z|[+-]\d{2}:\d{2}) \[\s{0,6}\d{1,}\] \[(first|second)\s{4,5}\] \[(INFO |ERROR)\] (first|second)$`)
	for _, line := range lines[4:] {
		if !pattern.MatchString(line) {
			t.Fatalf("malformed log line = %q", line)
		}
	}
}

// TestNewLoggerRemovesOnlyMatchingLogsOlderThanRetention verifies retention cleanup is scoped.
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

	logger, err := NewLogger(dir, 14, "local", "test")
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	defer logger.Close()

	data, err := os.ReadFile(filepath.Join(dir, logFileName(time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local))))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), "log retention cleanup triggered") || !strings.Contains(string(data), "removed expired log "+files[0]) {
		t.Fatalf("cleanup log = %q, want trigger and removed file", string(data))
	}

	assertMissing(t, filepath.Join(dir, logFileName(time.Date(2026, 6, 20, 12, 0, 0, 0, time.Local))))
	for _, name := range files[1:] {
		assertExists(t, filepath.Join(dir, name))
	}
}

// TestNewLoggerRejectsNegativeRetention verifies negative retention is invalid.
func TestNewLoggerRejectsNegativeRetention(t *testing.T) {
	_, err := NewLogger(filepath.Join(t.TempDir(), "logs"), -1, "local", "test")
	if err == nil {
		t.Fatal("expected NewLogger to reject negative retention")
	}
}

// TestEnsureLogFileOpenTreatsCleanupFailureAsWarning verifies housekeeping does not prevent logging.
func TestEnsureLogFileOpenTreatsCleanupFailureAsWarning(t *testing.T) {
	dir := t.TempDir()
	currentTime := time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local)
	freezeTime(t, currentTime)
	logger := &Logger{
		dir:           dir,
		retentionDays: -1,
		location:      time.Local,
		title:         "test",
	}
	defer logger.Close()

	if err := logger.ensureLogFileOpen(currentTime); err != nil {
		t.Fatalf("ensureLogFileOpen returned cleanup error: %v", err)
	}
	logger.Info("still available")

	data, err := os.ReadFile(filepath.Join(dir, logFileName(currentTime)))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), "[WARN ] log retention cleanup: log retention days must be non-negative") {
		t.Fatalf("log = %q, want cleanup warning", string(data))
	}
	if !strings.Contains(string(data), "still available") {
		t.Fatalf("log = %q, want subsequent entry", string(data))
	}
}

// TestLoggerWritesToStandardErrorWhenLogFileCannotOpen verifies writes do not use a stale or nil file.
func TestLoggerWritesToStandardErrorWhenLogFileCannotOpen(t *testing.T) {
	currentTime := time.Date(2026, 7, 6, 12, 0, 0, 0, time.Local)
	freezeTime(t, currentTime)
	var output bytes.Buffer
	oldStandardError := standardError
	standardError = &output
	t.Cleanup(func() { standardError = oldStandardError })

	logger := &Logger{
		dir:      filepath.Join(t.TempDir(), "missing", "logs"),
		location: time.Local,
		title:    "test",
	}
	logger.Error("not lost")

	if !strings.Contains(output.String(), "write log entry: open log file:") {
		t.Fatalf("stderr = %q, want open failure", output.String())
	}
	if !strings.Contains(output.String(), "not lost") {
		t.Fatalf("stderr = %q, want original entry", output.String())
	}
}

// TestLoggerUsesConfiguredTimezoneForFileAndEntries verifies time zone affects names and entries.
func TestLoggerUsesConfiguredTimezoneForFileAndEntries(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	freezeTime(t, time.Date(2026, 7, 6, 6, 30, 0, 0, time.UTC))

	logger, err := NewLogger(dir, 14, "America/Los_Angeles", "monitor")
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	defer logger.Close()

	logger.Info("hello")

	data, err := os.ReadFile(filepath.Join(dir, logFileName(time.Date(2026, 7, 5, 23, 30, 0, 0, time.FixedZone("PDT", -7*60*60)))))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.HasPrefix(string(data), "2026-07-05T23:30:00.000-07:00") {
		t.Fatalf("log line = %q, want configured timezone timestamp", string(data))
	}
}

// TestNewLoggerRejectsInvalidTimezone verifies bad time zone names fail construction.
func TestNewLoggerRejectsInvalidTimezone(t *testing.T) {
	_, err := NewLogger(filepath.Join(t.TempDir(), "logs"), 14, "Not/AZone", "test")
	if err == nil {
		t.Fatal("expected NewLogger to reject invalid timezone")
	}
}

// TestLoggerRotatesFilesAndRunsRetentionWhenDateChanges verifies midnight rotation and cleanup.
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

	logger, err := NewLogger(dir, 1, "local", "monitor")
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

// TestLoggerRotationRemovesExpiredOpenLog verifies cleanup can remove the previously active log.
func TestLoggerRotationRemovesExpiredOpenLog(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	currentTime := time.Date(2026, 7, 1, 12, 0, 0, 0, time.Local)
	oldNow := now
	now = func() time.Time { return currentTime }
	t.Cleanup(func() { now = oldNow })

	logger, err := NewLogger(dir, 2, "local", "monitor")
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	defer logger.Close()

	logger.Info("old entry")
	expiredLog := filepath.Join(dir, logFileName(currentTime))
	currentTime = currentTime.AddDate(0, 0, 3)
	logger.Info("new entry")

	assertMissing(t, expiredLog)
	data, err := os.ReadFile(filepath.Join(dir, logFileName(currentTime)))
	if err != nil {
		t.Fatalf("read current log: %v", err)
	}
	if !strings.Contains(string(data), "new entry") {
		t.Fatalf("current log = %q, want new entry", string(data))
	}
}

// freezeTime replaces now with value until the test cleanup runs.
func freezeTime(t *testing.T, value time.Time) {
	t.Helper()
	oldNow := now
	now = func() time.Time { return value }
	t.Cleanup(func() { now = oldNow })
}

// assertExists fails the test when path does not exist.
func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", filepath.Base(path), err)
	}
}

// assertMissing fails the test when path exists.
func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be removed, stat err = %v", filepath.Base(path), err)
	}
}
