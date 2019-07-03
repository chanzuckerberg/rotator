package sink_test

import (
	"os"
	"testing"

	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/shuheiktgw/go-travis"
	"github.com/stretchr/testify/require"
)

var (
	baseURL     = travis.ApiComUrl
	travisToken = os.Getenv("TRAVIS_API_AUTH_TOKEN")
	repoSlug    = "chanzuckerberg/shared-infra"
)

func TestWriteToTravisCiSink_Integration(t *testing.T) {
	r := require.New(t)

	client := travis.NewClient(baseURL, "")
	err := client.Authentication.UsingTravisToken(travisToken)
	// TODO: Or use github authentication (note: UsingGithubToken() calls UsingTravisToken() eventually anyway)
	// gitHubToken = os.Getenv("TRAVIS_GITHUB_PERSONAL_ACCESS_TOKEN")
	// travisToken, resp, err := client.Authentication.UsingGithubToken(ctx, gitHubToken)
	r.Nil(err)

	sink := sink.TravisCiSink{
		RepoSlug: repoSlug,
		Client:   client,
	}

	// create a key
	creds, err := (&source.DummySource{}).Read()
	r.Nil(err)
	err = sink.Write(creds)
	r.Nil(err)

	// update the key
	creds, err = (&source.DummySource{}).Read()
	r.Nil(err)
	err = sink.Write(creds)
	r.Nil(err)
}
