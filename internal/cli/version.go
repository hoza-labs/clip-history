package cli

import (
	"fmt"

	"github.com/hoza-labs/clip-history/internal/version"
)

// currentVersion is replaceable in tests and returns the version string to print.
var currentVersion = version.Current

// runVersionCommand prints the current version and returns nil.
func runVersionCommand() error {
	fmt.Println(currentVersion())
	return nil
}
