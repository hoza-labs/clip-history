// Package version resolves the application version for dev and release builds.
//
// Protocol: release builds inject buildVersion with -ldflags; development builds
// read the nearest VERSION file and append "-rc".
package version

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Current returns the injected release version or the VERSION file value with an "-rc" suffix.
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

// findVersionFile walks upward from the current directory and returns the trimmed VERSION contents.
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

// buildVersion is injected by release builds with -ldflags; an empty value means development mode.
var buildVersion string
