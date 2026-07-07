// Package safety decides whether clipboard text should be persisted.
//
// Protocol: call ShouldSave before writing clipboard text to history unless the
// user has explicitly enabled sensitive content storage.
package safety

import "strings"

// ShouldSave reports whether text is non-empty and does not contain obvious secret markers.
func ShouldSave(text string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}

	lower := strings.ToLower(text)
	for _, pattern := range []string{"password", "passwd", "secret", "token", "api key", "authorization", "private key"} {
		if strings.Contains(lower, pattern) {
			return false
		}
	}

	return true
}
