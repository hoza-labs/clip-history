package cli

import (
	"fmt"

	"github.com/hoza-labs/clip-history/internal/version"
)

var currentVersion = version.Current

func runVersionCommand() error {
	fmt.Println(currentVersion())
	return nil
}
