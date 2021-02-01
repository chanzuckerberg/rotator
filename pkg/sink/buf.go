package sink

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// A BufSink represents a sink that prints to a buffer.
type BufSink struct {
	BaseSink `yaml:",inline"`

	buf *bytes.Buffer
}

func NewBufSink() *BufSink {
	b := bytes.NewBuffer(nil)
	return &BufSink{buf: b}
}

func (sink *BufSink) WithKeyToName(m map[string]string) *BufSink {
	sink.BaseSink = BaseSink{KeyToName: m}
	return sink
}

func (sink *BufSink) Read() string {
	return sink.buf.String()
}

func (sink *BufSink) Write(ctx context.Context, name string, val interface{}) error {
	switch writeVal := val.(type) {
	case string:
		_, err := fmt.Fprint(sink.buf, writeVal)
		return errors.Wrap(err, "unable to write secret to buffer")
	default:
		return errors.Errorf("Buf Sink doesn't support writing type %T", writeVal)
	}
}

func (sink *BufSink) Kind() Kind {
	return KindBuf
}
