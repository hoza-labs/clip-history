package cli

import (
	"fmt"

	"github.com/hoza-labs/clip-history/internal/version"
)

func runVersionCommand() error {
	fmt.Println(version.Current())
	return nil
}
