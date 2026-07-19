// Package logging writes date-stamped command logs with retention cleanup.
//
// Protocol: create a Logger with NewLogger, check the error, then call
// defer logger.Close() immediately so the underlying file handle is released.
// Logger methods are safe for concurrent use and ignore writes after Close.
package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

// Logger writes timestamped entries to a daily log file.
type Logger struct {
	dir           string
	retentionDays int
	file          *os.File
	currentDay    time.Time
	location      *time.Location
	title         string
	mu            sync.Mutex
	closed        bool
}

// NewLogger opens a Logger aimed at directory `dir`; callers must call Close, usually with defer.
func NewLogger(dir string, retentionDays int, timezone string, title string) (*Logger, error) {
	if dir == "" {
		return nil, fmt.Errorf("log directory is required")
	}
	if retentionDays < 0 {
		return nil, fmt.Errorf("log retention days must be non-negative")
	}
	location, err := resolveLocation(timezone)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}
	logger := &Logger{
		dir:           dir,
		retentionDays: retentionDays,
		location:      location,
		title:         title,
	}
	if err := logger.ensureLogFileOpen(now().In(location)); err != nil {
		return nil, err
	}

	return logger, nil
}

// Info writes message at INFO level.
func (l *Logger) Info(message string) {
	l.write(levelInfo, message)
}

// Error writes message at ERROR level.
func (l *Logger) Error(message string) {
	l.write(levelError, message)
}

// Debug writes message at DEBUG level.
func (l *Logger) Debug(message string) {
	l.write(levelDebug, message)
}

// Close releases the current log file and returns any close error.
func (l *Logger) Close() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		l.closed = true
		return nil
	}
	err := l.file.Close()
	l.file = nil
	l.closed = true
	return err
}

// write appends message at level, appending to the log file appropriate for the current time
// and incidentally causing a "best effort" at removing old logs.
func (l *Logger) write(level level, message string) {
	if l == nil {
		return
	}
	location := l.location
	if location == nil {
		location = time.Local
	}
	currentTime := now().In(location)
	line := l.formatLogLine(currentTime, level, message)

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return
	}
	if err := l.ensureLogFileOpen(currentTime); err != nil {
		_, _ = io.WriteString(standardError, l.formatLogLine(currentTime, levelError, fmt.Sprintf("write log entry: %v", err)))
		_, _ = io.WriteString(standardError, line)
		return
	}
	_, _ = l.file.WriteString(line)
}

// ensureLogFileOpen opens the log file for at's local day and best-effort removes expired logs.
func (l *Logger) ensureLogFileOpen(at time.Time) error {
	day := beginningOfDay(at.In(l.location))
	if day.Equal(l.currentDay) {
		return nil
	}

	path := filepath.Join(l.dir, logFileName(day))
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	// Close the previous file before cleanup so it can be removed when the logger
	// has been idle past the retention period.
	oldFile := l.file
	l.file = file
	l.currentDay = day
	if oldFile != nil {
		_ = oldFile.Close()
	}

	// The local callback reports cleanup work without re-entering Logger.write.
	logInfo := func(message string) {
		_, _ = file.WriteString(l.formatLogLine(now(), levelInfo, message))
	}
	if err := removeExpiredLogs(l.dir, l.retentionDays, l.location, logInfo); err != nil {
		_, _ = file.WriteString(l.formatLogLine(now(), levelWarn, fmt.Sprintf("log retention cleanup: %v", err)))
	}

	return nil
}

// level identifies the severity of a log entry.
type level string

const (
	// levelInfo identifies an informational entry.
	levelInfo level = "INFO"
	// levelError identifies an error entry.
	levelError level = "ERROR"
	// levelWarn identifies a warning entry.
	levelWarn level = "WARN"
	// levelDebug identifies a diagnostic entry.
	levelDebug level = "DEBUG"
)

// logTimestampLayout formats timestamps written at the start of each log entry.
const logTimestampLayout = "2006-01-02T15:04:05.000Z07:00"

// formatLogLine returns one consistently formatted entry using l's title and location.
func (l *Logger) formatLogLine(at time.Time, level level, message string) string {
	return fmt.Sprintf(
		"%s [%7d] [%-10s] [%-5s] %s\n",
		at.In(l.location).Format(logTimestampLayout), os.Getpid(), l.title, level, message)
}

const (
	// logFileDateLayout formats the date embedded in log file names.
	logFileDateLayout = "2006-01-02"
	// logFilePrefix identifies log files managed by this package.
	logFilePrefix = "clip-history-"
	// logFileSuffix identifies the file type used for managed logs.
	logFileSuffix = ".log"
)

// logFilePattern matches managed clip-history log files and captures their date.
var logFilePattern = regexp.MustCompile(`^` + regexp.QuoteMeta(logFilePrefix) + `(\d{4}-\d{2}-\d{2})` + regexp.QuoteMeta(logFileSuffix) + `$`)

// logFileName returns the managed log file name for at's date.
func logFileName(at time.Time) string {
	return logFilePrefix + at.Format(logFileDateLayout) + logFileSuffix
}

// removeExpiredLogs deletes managed logs older than retentionDays in location and reports its decisions through logInfo.
func removeExpiredLogs(dir string, retentionDays int, location *time.Location, logInfo func(string)) error {
	logInfo("log retention cleanup triggered")
	if retentionDays < 0 {
		return fmt.Errorf("log retention days must be non-negative")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read log dir: %w", err)
	}

	cutoff := beginningOfDay(now().In(location).AddDate(0, 0, -retentionDays))
	removed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := logFilePattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		logDate, err := time.ParseInLocation(logFileDateLayout, matches[1], location)
		if err != nil {
			continue
		}
		if !logDate.Before(cutoff) {
			continue
		}
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
			return fmt.Errorf("remove expired log %s: %w", entry.Name(), err)
		}
		removed++
		logInfo(fmt.Sprintf("removed expired log %s", entry.Name()))
	}
	if removed == 0 {
		logInfo("log retention cleanup complete; no expired logs found")
	} else {
		logInfo(fmt.Sprintf("log retention cleanup complete; removed %d expired log(s)", removed))
	}

	return nil
}

// resolveLocation returns the configured time zone, treating "" and "local" as time.Local.
func resolveLocation(timezone string) (*time.Location, error) {
	if timezone == "" || timezone == "local" {
		return time.Local, nil
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("load log timezone %q: %w", timezone, err)
	}
	return location, nil
}

// beginningOfDay returns t truncated to midnight in t's location.
func beginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// now is replaceable in tests and returns the current time.
var now = time.Now

// standardError receives entries that cannot be written to their intended log file.
var standardError io.Writer = os.Stderr
