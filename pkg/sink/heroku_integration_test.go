package sink

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	heroku "github.com/heroku/heroku-go/v5"
	"github.com/stretchr/testify/require"
)

func TestWriteToHerokuSink_Integration(t *testing.T) {
	r := require.New(t)

	// create a client (hopefully covered by herokutestsuite)
	testHeader := http.Header{}
	testHeader.Set("Accept", "application/vnd.heroku+json; version=3") // should I make these references more
	testTransport := heroku.Transport{
		AdditionalHeaders: testHeader,
		Debug:             true,
	}
	heroku.DefaultClient.Transport = &testTransport
	herokuService := heroku.NewService(heroku.DefaultClient)

	// Do I mock a http server & its response to GET requests? What do I return???

	// create a sink
	sink := HerokuSink{
		AppIdentity: "testIdentity",
		Client:      herokuService,
	}
	r.NotNil(sink.Client)

	oldSecret := "oldValue"
	testEnvVarMap := map[string]*string{
		"testEnv": aws.String(oldSecret),
	}
	sink.Client.ConfigVarUpdate(context.TODO(), "testIdentity", testEnvVarMap)

	//
	configVarUpdate, err := sink.Client.ConfigVarInfoForApp(context.TODO(), sink.AppIdentity)
	r.NoError(err)
	r.NotNil(configVarUpdate["testEnv"])
	r.Equal(*configVarUpdate["testEnv"], oldSecret)

	// create new secret?
	NewSecret := "newValue"

	// rotate secret by creating a mock config (with sink.AppIdentity = "testIdentity")
	// run RotateSecrets() from cmd/rotate.go and see if the secret got rotated

	// check that the secret value is updated
	r.NotNil(configVarUpdate["testEnv"])
	r.Equal(*configVarUpdate["testEnv"], NewSecret)

}
