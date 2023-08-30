package logr

import (
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/go-logr/logr"
)

var _ io.Writer = &LogWriter{}

type LogWriter struct {
	log        logr.Logger
	buffer     bytes.Buffer
	lineBuffer bytes.Buffer
}

func NewLogWriter(log logr.Logger) *LogWriter {
	return &LogWriter{
		log: log,
	}
}

// Write implements io.Writer.
func (w *LogWriter) Write(p []byte) (n int, err error) {
	n, err = w.buffer.Write(p)
	if err != nil {
		return
	}
	for {
		b, err := w.buffer.ReadBytes(byte('\n'))
		if _, err := w.lineBuffer.Write(b); err != nil {
			return n, err
		}
		if err != nil {
			return n, nil
		}
		writeLine := strings.TrimSpace(StripAnsi(w.lineBuffer.String()))
		w.log.Info(writeLine)
		w.lineBuffer.Reset()
	}
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func StripAnsi(str string) string {
	return re.ReplaceAllString(str, "")
}
