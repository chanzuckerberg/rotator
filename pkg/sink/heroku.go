package sink

import (
	"context"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	heroku "github.com/heroku/heroku-go/v5"
	"github.com/pkg/errors"
)

type HerokuSink struct {
	BaseSink    `yaml:",inline"`
	Client      *heroku.Service `yaml:"client"`
	AppIdentity string          `yaml:"AppIdentity"`
}

func (sink *HerokuSink) WithHerokuClient(client *heroku.Service) *HerokuSink {
	sink.Client = client
	return sink
}

// Write writes the value of the env var with the specified name for the given repo
func (sink *HerokuSink) Write(ctx context.Context, name string, val string) error {
	keypair := map[string]*string{
		name: &val,
	}
	if sink.Client == nil {
		return errors.New("Heroku Client not set")
	}
	updateResult, err := sink.Client.ConfigVarUpdate(ctx, sink.AppIdentity, keypair)
	if err != nil {
		return errors.Wrapf(err, "Unable to update Config var with %s:%s", name, val)
	}
	fmt.Println("updateResult:")
	spew.Dump(updateResult)
	fmt.Printf("sink:Heroku: \n name: %s, val: %#v\n", name, val)
	return nil
}

// Kind returns the kind of this sink
func (sink *HerokuSink) Kind() Kind {
	return KindHeroku
}
