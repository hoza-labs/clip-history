package safety

import "testing"

func TestShouldSaveRejectsSensitiveText(t *testing.T) {
	if ShouldSave("my password is secret") {
		t.Fatal("expected sensitive text to be rejected")
	}
}

func TestShouldSaveAcceptsNormalText(t *testing.T) {
	if !ShouldSave("plain clipboard text") {
		t.Fatal("expected ordinary text to be saved")
	}
}
