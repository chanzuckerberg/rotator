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

func (sink *BufSink) Write(creds map[string]string) error {
	for _, v := range creds {
		_, err := fmt.Fprint(sink.buf, v)
		if err != nil {
			return errors.Wrap(err, "unable to write secret to buffer")
		}
	}
	return nil
}

func (sink *BufSink) Kind() Kind {
	return KindBuf
}
