package sender

import (
	"fmt"
	"io"
	"os"
	"time"
)

// LogSender writes alert messages to a configured writer (default: stdout).
type LogSender struct {
	Out io.Writer
}

// NewLogSender returns a LogSender that writes to stdout.
func NewLogSender() *LogSender {
	return &LogSender{Out: os.Stdout}
}

// Send writes a formatted alert line to the configured writer.
func (l *LogSender) Send(level, message string) error {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	_, err := fmt.Fprintf(l.Out, "%s [%s] %s\n", timestamp, level, message)
	return err
}
