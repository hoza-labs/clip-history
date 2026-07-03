package version

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCurrentReadsVersionFile(t *testing.T) {
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	defer func() {
		_ = os.Chdir(oldDir)
		buildVersion = ""
	}()

	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte("0.0.1\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := os.Chdir(nested); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	if got := Current(); got != "0.0.1" {
		t.Fatalf("Current() = %q, want %q", got, "0.0.1")
	}
}

func TestCurrentPrefersBuildVersion(t *testing.T) {
	buildVersion = "1.2.3"
	t.Cleanup(func() { buildVersion = "" })

	if got := Current(); got != "1.2.3" {
		t.Fatalf("Current() = %q, want %q", got, "1.2.3")
	}
}
