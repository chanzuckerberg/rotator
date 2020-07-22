package sink

import (
	"context"
)

type HerokuSink struct {
	owner     string
	repo      string
	KeyToName map[string]string
}

func (sink *HerokuSink) GetKeyToName() map[string]string {
	return sink.KeyToName
}

// Write writes the value of the env var with the specified name for the given repo
func (sink *HerokuSink) Write(ctx context.Context, name string, val string) error {
	sink.KeyToName[name] = val
	return nil
}

// Kind returns the kind of this sink
func (sink *HerokuSink) Kind() Kind {
	return KindHeroku
}
