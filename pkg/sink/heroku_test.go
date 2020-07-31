package sink

import (
	"net/http/httptest"
	"testing"

	heroku "github.com/heroku/heroku-go/v5"
	"github.com/stretchr/testify/require"
)

const (
	herokuEnvVar    = "foo"
	herokuEnvVarVal = "bar"
)

func TestHerokuSinkWrite(t *testing.T) {
	r := require.New(t)
	r.Nil(nil)

	testServer := httptest.NewServer(nil)
	defer testServer.Close()

	hs := heroku.NewService(nil)
	hs.URL = testServer.URL
}
