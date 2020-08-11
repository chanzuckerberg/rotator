package sink

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	heroku "github.com/heroku/heroku-go/v5"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"
)

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
func TestHerokuSinkWrite(t *testing.T) {
	r := require.New(t)
	r.Nil(nil)

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

	testSink := HerokuSink{
		AppIdentity: "testIdentity",
		Client:      herokuService,
	}
	r.NotNil(testSink.Client)

	oldSecret := "oldValue"
	testEnvVarMap := map[string]*string{
		"testEnv": heroku.String(oldSecret),
	}
	testSink.Client.ConfigVarUpdate(context.TODO(), "testIdentity", testEnvVarMap)

	configVarUpdate, err := testSink.Client.ConfigVarInfoForApp(context.TODO(), testSink.AppIdentity)
	r.NoError(err)
	r.NotNil(configVarUpdate["testEnv"])
	r.Equal(oldSecret, *configVarUpdate["testEnv"])

}