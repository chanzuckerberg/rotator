package sink

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
)

// A StdoutSink represents a sink that prints to a buffer.
type BufSink struct {
	buf *bytes.Buffer
}

func NewBufSink() *BufSink {
	b := bytes.NewBuffer(nil)
	return &BufSink{buf: b}
}

func (sink *BufSink) Read() string {
	return sink.buf.String()
}

// Write writes secret to sink.buf.
func (sink *BufSink) Write(secret string) error {
	_, err := fmt.Fprint(sink.buf, secret)
	return errors.Wrap(err, "unable to write secret to buffer")
}

func (sink *BufSink) Kind() Kind {
	return KindBuf
}
