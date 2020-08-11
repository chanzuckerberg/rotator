package config_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/chanzuckerberg/rotator/cmd"
	"github.com/chanzuckerberg/rotator/pkg/config"
	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/chanzuckerberg/rotator/pkg/util"
	heroku "github.com/heroku/heroku-go/v5"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestFromFile(t *testing.T) {
	r := require.New(t)

	tmpFile, err := ioutil.TempFile("", "tmpConfig")
	r.Nil(err)
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	c1 := &config.Config{Secrets: []config.Secret{}}
	bytes, err := yaml.Marshal(c1)
	r.Nil(err)
	_, err = tmpFile.Write(bytes)
	r.Nil(err)

	c2, err := config.FromFile(tmpFile.Name())
	r.Nil(err)

	r.Equal(c1, c2)
}

type authorizedHerokuClient struct {
	Transport     heroku.Transport
	configVarInfo map[string]*string
	URL           string
}

func (c *authorizedHerokuClient) ConfigVarUpdate(ctx context.Context, appIdentity string, o map[string]*string) (heroku.ConfigVarUpdateResult, error) {
	for k, v := range o {
		c.configVarInfo[k] = v
	}
	// TODO: verify if a nil updateresult means the update was successful. We don't need ths output for this file's test
	return nil, nil
}

func (c *authorizedHerokuClient) ConfigVarInfoForApp(ctx context.Context, appIdentity string) (heroku.ConfigVarInfoForAppResult, error) {
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

func TestRotate(t *testing.T) {
	r := require.New(t)

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
	herokuService := authorizedHerokuClient{
		Transport:     heroku.Transport{},
		configVarInfo: map[string]*string{},
		URL:           testServer.URL,
	}

	testSink := sink.HerokuSink{
		AppIdentity: "testIdentity",
		Client:      herokuService,
	}

	NewSecret := "newValue"
	defer util.ResetEnv(os.Environ())
	err := os.Setenv("HEROKU_BEARER_TOKEN", NewSecret)
	r.NoError(err)

	testHerokuSinkConfig := &config.Config{
		Version: 0,
		Secrets: []config.Secret{
			{
				Name: "test",
				Source: &source.Env{
					Name: "TEST_ENV",
				},
				Sinks: []sink.Sink{
					testSink,
				},
			},
		},
	}
	// rotate secret by creating a mock config (with sink.AppIdentity = "testIdentity")
	// run RotateSecrets() from cmd/rotate.go and see if the secret got rotated
	cmd.RotateSecrets(testHerokuSinkConfig)
	// check that the secret value is updated
	configVarUpdate, err := testSink.Client.ConfigVarInfoForApp(context.TODO(), testSink.AppIdentity)
	r.NotNil(configVarUpdate)
	r.NotNil(configVarUpdate["testEnv"])
	r.Equal(NewSecret, *configVarUpdate["testEnv"])
}
