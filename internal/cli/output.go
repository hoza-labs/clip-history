package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// commandOutput is the user-visible output behavior required by command handlers.
type commandOutput interface {
	// Printf writes formatted output.
	Printf(format string, args ...any)
	// Println writes output followed by a newline.
	Println(args ...any)
}

// loggedOutput writes command output to stdout and mirrors non-empty lines to the command log.
type loggedOutput struct {
	writer io.Writer
	logger commandLogger
}

// newLoggedOutput returns stdout-backed output that also logs through logger when logger is non-nil.
func newLoggedOutput(logger commandLogger) loggedOutput {
	return loggedOutput{
		writer: os.Stdout,
		logger: logger,
	}
}

// Printf writes fmt.Sprintf(format, args...) to the output sink and log.
func (o loggedOutput) Printf(format string, args ...any) {
	text := fmt.Sprintf(format, args...)
	o.write(text)
}

// Println writes fmt.Sprintln(args...) to the output sink and log.
func (o loggedOutput) Println(args ...any) {
	text := fmt.Sprintln(args...)
	o.write(text)
}

// write sends text to the configured writer and logs the trimmed text when it is non-empty.
func (o loggedOutput) write(text string) {
	writer := o.writer
	if writer == nil {
		writer = os.Stdout
	}
	_, _ = io.WriteString(writer, text)

	if o.logger == nil {
		return
	}
	message := strings.TrimRight(text, "\r\n")
	if message == "" {
		return
	}
	o.logger.Info("> " + message)
}
