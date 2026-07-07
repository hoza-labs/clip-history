package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type commandOutput interface {
	Printf(format string, args ...any)
	Println(args ...any)
}

type loggedOutput struct {
	writer io.Writer
	logger commandLogger
}

func newLoggedOutput(logger commandLogger) loggedOutput {
	return loggedOutput{
		writer: os.Stdout,
		logger: logger,
	}
}

func (o loggedOutput) Printf(format string, args ...any) {
	text := fmt.Sprintf(format, args...)
	o.write(text)
}

func (o loggedOutput) Println(args ...any) {
	text := fmt.Sprintln(args...)
	o.write(text)
}

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
