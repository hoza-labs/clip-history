package safety

import "testing"

// TestShouldSaveRejectsSensitiveText verifies obvious secret markers are filtered.
func TestShouldSaveRejectsSensitiveText(t *testing.T) {
	if ShouldSave("my password is secret") {
		t.Fatal("expected sensitive text to be rejected")
	}
}

// TestShouldSaveAcceptsNormalText verifies ordinary clipboard text is allowed.
func TestShouldSaveAcceptsNormalText(t *testing.T) {
	if !ShouldSave("plain clipboard text") {
		t.Fatal("expected ordinary text to be saved")
	}
}
