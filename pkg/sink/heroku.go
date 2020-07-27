package sink

import (
	"context"
	"fmt"
)

type HerokuSink struct {
	owner     string
	repo      string
	BaseSink  `yaml:",inline"`
	KeyToName map[string]string
	EnvVars   map[string]string
}

func (sink *HerokuSink) WithKeyToName(m map[string]string) *HerokuSink {
	sink.BaseSink = BaseSink{KeyToName: m}
	return sink
}

// Write writes the value of the env var with the specified name for the given repo
func (sink *HerokuSink) Write(ctx context.Context, name string, val string) error {
	// make a map of existing env vars

	// find env var by name
	fmt.Printf("sink:Heroku: \n name: %s, val: %#v\n", name, val)
	// update
	sink.KeyToName[name] = val
	return nil
}

// Kind returns the kind of this sink
func (sink *HerokuSink) Kind() Kind {
	return KindHeroku
}
