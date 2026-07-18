// Package logging writes date-stamped command logs with retention cleanup.
//
// Protocol: create a Logger with NewLogger, check the error, then call
// defer logger.Close() immediately so the underlying file handle is released.
// Logger methods are safe for concurrent use and ignore writes after Close.
package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

// logFileDateLayout formats the date embedded in log file names.
const logFileDateLayout = "2006-01-02"

// logTimestampLayout formats timestamps written at the start of each log entry.
const logTimestampLayout = "2006-01-02T15:04:05.000Z07:00"

var (
	// logFilePattern matches managed clip-history log files and captures their date.
	logFilePattern = regexp.MustCompile(`^clip-history-(\d{4}-\d{2}-\d{2})\.log$`)
	// now is replaceable in tests and returns the current time.
	now = time.Now
)

// Logger writes timestamped entries to a daily log file.
type Logger struct {
	dir           string
	retentionDays int
	file          *os.File
	currentDate   string
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
	if err := logger.rotate(now().In(location).Format(logFileDateLayout)); err != nil {
		return nil, err
	}

	return logger, nil
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

// beginningOfDay returns t truncated to midnight in t's location.
func beginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// Info writes message at INFO level.
func (l *Logger) Info(message string) {
	l.write("INFO", message)
}

// Error writes message at ERROR level.
func (l *Logger) Error(message string) {
	l.write("ERROR", message)
}

// Debug writes message at DEBUG level.
func (l *Logger) Debug(message string) {
	l.write("DEBUG", message)
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

// write appends message at level, rotating the log file when the local date changes.
func (l *Logger) write(level, message string) {
	if l == nil {
		return
	}
	location := l.location
	if location == nil {
		location = time.Local
	}
	currentTime := now().In(location)
	timestamp := currentTime.Format(logTimestampLayout)
	date := currentTime.Format(logFileDateLayout)
	line := fmt.Sprintf("%s [%7d] [%-10s] [%-5s] %s\n", timestamp, os.Getpid(), l.title, level, message)

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return
	}
	if date != l.currentDate {
		if err := l.rotate(date); err != nil {
			_, _ = l.file.WriteString(fmt.Sprintf("%s [%7d] [%-10s] [ERROR] rotate log file: %v\n", timestamp, os.Getpid(), l.title, err))
		}
	}
	_, _ = l.file.WriteString(line)
}

// rotate opens the log file for date and closes the previous file if one was open.
func (l *Logger) rotate(date string) error {
	path := filepath.Join(l.dir, fmt.Sprintf("clip-history-%s.log", date))
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	// Keep the previous file so a cleanup failure can restore it, or close it
	// after rotation succeeds. The local logInfo callback lets retention cleanup
	// report its work directly to the new file without re-entering Logger.write.
	oldFile := l.file
	l.file = file
	l.currentDate = date
	logInfo := func(message string) {
		currentTime := now().In(l.location)
		_, _ = l.file.WriteString(fmt.Sprintf("%s [%7d] [%-10s] [INFO ] %s\n", currentTime.Format(logTimestampLayout), os.Getpid(), l.title, message))
	}
	if err := removeExpiredLogs(l.dir, l.retentionDays, l.location, logInfo); err != nil {
		l.file = oldFile
		_ = file.Close()
		return err
	}
	if oldFile != nil {
		_ = oldFile.Close()
	}

	return nil
}
