package sink

import (
	"context"

	heroku "github.com/heroku/heroku-go/v5"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

func NewHerokuSink() *HerokuSink {
	return &HerokuSink{}
}

func (sink *HerokuSink) WithHerokuClient(client HerokuServiceIface) *HerokuSink {
	sink.Client = client
	return sink
}

// Write writes the value of the env var with the specified name for the given repo
func (sink *HerokuSink) Write(ctx context.Context, name string, val string) error {

	varUpdates := map[string]*string{
		name: &val,
	}
	if sink.Client == nil {
		return errors.New("Heroku Client not set")
	}
	if sink.AppIdentity == "" {
		return errors.New("Heroku AppIdentity not set")
	}

	_, err := sink.Client.ConfigVarUpdate(ctx, sink.AppIdentity, varUpdates)
	if err != nil {
		return errors.Wrapf(err, "Unable to update Config var with %s:%s", name, val)
	}

	logrus.Debugf("sink:Heroku: \n name: %s, val: %#v\n", name, val)
	return nil
}

// Kind returns the kind of this sink
func (sink *HerokuSink) Kind() Kind {
	return KindHeroku
}
