package version

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// released version will permanently inject a value here, via -ldflags; cf create-release.sh
var buildVersion string

func Current() string {
	if value := strings.TrimSpace(buildVersion); value != "" {
		return value
	}

	value, err := findVersionFile()
	if err != nil {
		return fmt.Sprintf("dev (%v)", err)
	}
	if value == "" {
		return "dev"
	}

	return value + "-rc"
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
