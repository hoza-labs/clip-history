package main

import (
	"fmt"
	"io"
	"os"

	"github.com/hoza-labs/clip-history/internal/cli"
)

func main() {
	run(cli.Execute, os.Stderr, os.Exit)
}

func run(execute func() error, stderr io.Writer, exit func(int)) {
	if err := execute(); err != nil {
		fmt.Fprintln(stderr, err)
		exit(1)
	}
}
