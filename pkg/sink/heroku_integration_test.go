package sink

import (
	"context"
	"testing"

	heroku "github.com/heroku/heroku-go/v5"
	"github.com/stretchr/testify/require"
)

func TestWriteToHerokuSink_Integration(t *testing.T) {
	r := require.New(t)

	// create a client (hopefully covered by herokutestsuite)
	HerokuClient := heroku.NewService(nil)
	// create a sink
	sink := HerokuSink{
		AppIdentity: "testIdentity",
		Client:      HerokuClient,
	}
	r.NotNil(sink.Client)

	// get original secret
	configVarUpdate, err := sink.Client.ConfigVarInfoForApp(context.TODO(), sink.AppIdentity)
	r.NoError(err)
	r.Equal(configVarUpdate["testEnv"], "oldValue")

	// create new secret?

	// rotate secret

	// check that the secret value is updated
	r.Equal(configVarUpdate["testEnv"], "newValue")

}
