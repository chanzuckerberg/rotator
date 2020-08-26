package sink

import (
	"context"
	"fmt"

	heroku "github.com/heroku/heroku-go/v5"
	"github.com/pkg/errors"
)

type HerokuServiceIface interface {
	ConfigVarUpdate(ctx context.Context, appIdentity string, o map[string]*string) (heroku.ConfigVarUpdateResult, error)
	ConfigVarInfoForApp(ctx context.Context, appIdentity string) (heroku.ConfigVarInfoForAppResult, error)
}

type HerokuSink struct {
	BaseSink    `yaml:",inline"`
	Client      HerokuServiceIface `yaml:"client"`
	AppIdentity string             `yaml:"AppIdentity"`
}

func (sink *HerokuSink) WithHerokuClient(client HerokuServiceIface) *HerokuSink {
	sink.Client = client
	return sink
}

func (sink *HerokuSink) GetKeyToName() map[string]string {
	return sink.KeyToName
}

func (sink *HerokuSink) WithKeyToName(m map[string]string) *HerokuSink {
	sink.BaseSink = BaseSink{KeyToName: m}
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
	if sink.AppIdentity == "" {
		return errors.New("Heroku AppIdentity not set")
	}

	_, err := sink.Client.ConfigVarUpdate(ctx, sink.AppIdentity, keypair)
	if err != nil {
		return errors.Wrapf(err, "Unable to update Config var with %s:%s", name, val)
	}

	fmt.Printf("sink:Heroku: \n name: %s, val: %#v\n", name, val)
	return nil
}

// Kind returns the kind of this sink
func (sink *HerokuSink) Kind() Kind {
	return KindHeroku
}
