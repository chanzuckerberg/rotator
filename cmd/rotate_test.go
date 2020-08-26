package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/chanzuckerberg/rotator/pkg/config"
	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/chanzuckerberg/rotator/pkg/util"
	heroku "github.com/heroku/heroku-go/v5"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"
)

type authorizedHerokuClient struct {
	Transport     heroku.Transport
	configVarInfo map[string]*string
	URL           string
}

func (c authorizedHerokuClient) ConfigVarUpdate(ctx context.Context, appIdentity string, o map[string]*string) (heroku.ConfigVarUpdateResult, error) {
	for k, v := range o {
		c.configVarInfo[k] = v
	}
	// TODO: verify if a nil updateresult means the update was successful. We don't need ths output for this file's test
	return nil, nil
}

func (c authorizedHerokuClient) ConfigVarInfoForApp(ctx context.Context, appIdentity string) (heroku.ConfigVarInfoForAppResult, error) {
	return c.configVarInfo, nil
}

func testAuthenticate() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}
}
func mockTestRouter(
	ctx context.Context,
) http.Handler {
	router := httprouter.New()
	handle := testAuthenticate()

	router.GET("/", handle)
	return router
}

func TestHerokuSinkWrite(t *testing.T) {
	r := require.New(t)
	r.Nil(nil)

	// set up mock Heroku router
	router := mockTestRouter(context.TODO())
	testServer := httptest.NewServer(router)
	defer testServer.Close()
	testHeader := http.Header{}
	testHeader.Set("Accept", "application/vnd.heroku+json; version=3")
	testTransport := heroku.Transport{
		AdditionalHeaders: testHeader,
		Debug:             true,
	}
	heroku.DefaultClient.Transport = &testTransport
	herokuService := &authorizedHerokuClient{
		Transport:     heroku.Transport{},
		configVarInfo: map[string]*string{},
		URL:           testServer.URL,
	}

	// Insert the old secret into the Heroku service
	oldSecret := "oldValue"
	secretName := "test_env"
	testSink := sink.HerokuSink{
		AppIdentity: "testIdentity",
		Client:      herokuService,
		BaseSink: sink.BaseSink{
			KeyToName: map[string]string{
				"TEST_ENV": secretName,
			},
		},
	}
	r.NotNil(testSink.Client)
	testEnvVarMap := map[string]*string{
		secretName: heroku.String(oldSecret),
	}
	testSink.Client.ConfigVarUpdate(context.TODO(), "testIdentity", testEnvVarMap)
	configVarUpdate, err := testSink.Client.ConfigVarInfoForApp(context.TODO(), testSink.AppIdentity)
	r.NoError(err)

	// Verify the old secret is in the heroku service's config vars
	r.NotNil(configVarUpdate[secretName])
	r.Equal(oldSecret, *configVarUpdate[secretName])

	// Set the environment variables (with new secret value) needed for the test
	NewSecret := "newValue"
	defer util.ResetEnv(os.Environ())
	err = os.Setenv("HEROKU_BEARER_TOKEN", NewSecret)
	r.NoError(err)
	err = os.Setenv("TEST_ENV", NewSecret)
	r.NoError(err)

	// Rotate secrets
	testHerokuSinkConfig := &config.Config{
		Version: 0,
		Secrets: []config.Secret{
			{
				Name: "test",
				Source: &source.Env{
					Name: "TEST_ENV",
				},
				Sinks: []sink.Sink{
					&testSink,
				},
			},
		},
	}
	RotateSecrets(testHerokuSinkConfig)

	// check that the secret value is updated
	configVarUpdate, err = testSink.Client.ConfigVarInfoForApp(context.TODO(), testSink.AppIdentity)
	r.NoError(err)
	r.NotNil(configVarUpdate)
	r.NotNil(configVarUpdate[secretName])
	r.Equal(NewSecret, *configVarUpdate[secretName])
}
