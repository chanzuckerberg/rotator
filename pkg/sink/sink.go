package sink

import "fmt"

type Sink interface {
	Write(secret string) error
}

type StdoutSink struct{}

// Writes its string input to stdout.
func (sink StdoutSink) Write(secret string) error {
	fmt.Println(secret)
	return nil
}
