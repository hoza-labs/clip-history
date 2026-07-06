package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Logger struct {
	file *os.File
	mu   sync.Mutex
}

func NewLogger(dir string) (Logger, error) {
	if dir == "" {
		return Logger{}, fmt.Errorf("log directory is required")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Logger{}, fmt.Errorf("create log dir: %w", err)
	}

	path := filepath.Join(dir, "clip-history.log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return Logger{}, fmt.Errorf("open log file: %w", err)
	}

	return Logger{file: file}, nil
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
	timestamp := time.Now().UTC().Format(time.RFC3339)
	line := fmt.Sprintf("%s [%7d] [%-5s] %s\n", timestamp, os.Getpid(), level, message)

	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.file.WriteString(line)
}
