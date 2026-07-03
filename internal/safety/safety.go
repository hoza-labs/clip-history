package safety

import "strings"

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
