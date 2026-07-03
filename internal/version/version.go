package version

import (
	"os"
	"path/filepath"
	"strings"
)

var buildVersion string

func Current() string {
	if value := strings.TrimSpace(buildVersion); value != "" {
		return value
	}

	value, err := findVersionFile()
	if err != nil {
		return "dev"
	}
	if value == "" {
		return "dev"
	}

	return value
}

func findVersionFile() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		path := filepath.Join(dir, "VERSION")
		content, readErr := os.ReadFile(path)
		if readErr == nil {
			return strings.TrimSpace(string(content)), nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", readErr
		}

		dir = parent
	}
}
