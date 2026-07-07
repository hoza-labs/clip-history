// Package main wires the ch executable to the internal CLI package.
//
// Protocol: main delegates all command parsing to cli.Execute and exits non-zero only
// after writing a returned error to stderr.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/hoza-labs/clip-history/internal/cli"
)

// main runs the clip-history command-line program.
func main() {
	run(cli.Execute, os.Stderr, os.Exit)
}

// run invokes execute, writes any returned error to stderr, and calls exit with code 1.
func run(execute func() error, stderr io.Writer, exit func(int)) {
	if err := execute(); err != nil {
		fmt.Fprintln(stderr, err)
		exit(1)
	}
}
