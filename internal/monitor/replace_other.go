//go:build !windows

package monitor

import "os"

// readStateFile reads all state data from path on non-Windows systems.
func readStateFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// replaceFile atomically moves source over destination on non-Windows systems.
func replaceFile(source, destination string) error {
	return os.Rename(source, destination)
}
