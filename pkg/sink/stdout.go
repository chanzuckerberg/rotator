package sink

import (
	"context"
	"fmt"
)

// A StdoutSink
type StdoutSink struct {
	BaseSink `yaml:",inline"`
}

func NewStdoutSink() *StdoutSink {
	return &StdoutSink{}
}

func (sink *StdoutSink) WithKeyToName(m map[string]string) *StdoutSink {
	sink.BaseSink = BaseSink{KeyToName: m}
	return sink
}

func (sink *StdoutSink) Write(ctx context.Context, name string, val string) error {
	fmt.Printf("sink:stdout: \n name: %s, val: %#v\n", name, val)
	return nil
}

func (sink *StdoutSink) Kind() Kind {
	return KindStdout
}
