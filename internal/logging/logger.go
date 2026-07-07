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

var (
	logFilePattern = regexp.MustCompile(`^clip-history-(\d{4}-\d{2}-\d{2})\.log$`)
	now            = time.Now
)

type Logger struct {
	file     *os.File
	location *time.Location
	mu       sync.Mutex
}

func NewLogger(dir string, retentionDays int, timezone string) (*Logger, error) {
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
	if err := removeExpiredLogs(dir, retentionDays, location); err != nil {
		return nil, err
	}

	path := filepath.Join(dir, fmt.Sprintf("clip-history-%s.log", now().In(location).Format(logFileDateLayout)))
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	return &Logger{file: file, location: location}, nil
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
	if l.file == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

func (l *Logger) write(level, message string) {
	if l == nil || l.file == nil {
		return
	}
	location := l.location
	if location == nil {
		location = time.Local
	}
	timestamp := now().In(location).Format(time.RFC3339)
	line := fmt.Sprintf("%s [%7d] [%-5s] %s\n", timestamp, os.Getpid(), level, message)

	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.file.WriteString(line)
}
