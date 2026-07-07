package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

const logFileDateLayout = "2006-01-02"
const logTimestampLayout = "2006-01-02T15:04:05.000Z07:00"

var (
	logFilePattern = regexp.MustCompile(`^clip-history-(\d{4}-\d{2}-\d{2})\.log$`)
	now            = time.Now
)

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

func removeExpiredLogs(dir string, retentionDays int, location *time.Location) error {
	if retentionDays < 0 {
		return fmt.Errorf("log retention days must be non-negative")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read log dir: %w", err)
	}

	cutoff := beginningOfDay(now().In(location).AddDate(0, 0, -retentionDays))
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
	}

	return nil
}

func beginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func (l *Logger) Info(message string) {
	l.write("INFO", message)
}

func (l *Logger) Error(message string) {
	l.write("ERROR", message)
}

func (l *Logger) Debug(message string) {
	l.write("DEBUG", message)
}

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

func (l *Logger) rotate(date string) error {
	if err := removeExpiredLogs(l.dir, l.retentionDays, l.location); err != nil {
		return err
	}

	path := filepath.Join(l.dir, fmt.Sprintf("clip-history-%s.log", date))
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	oldFile := l.file
	l.file = file
	l.currentDate = date
	if oldFile != nil {
		_ = oldFile.Close()
	}

	return nil
}
