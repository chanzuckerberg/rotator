package sink_test

import (
	"context"
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
	repoSlug    = "chanzuckerberg/rotator"
)

func TestWriteToTravisCiSink_Integration(t *testing.T) {
	r := require.New(t)

	client := travis.NewClient(baseURL, "")
	err := client.Authentication.UsingTravisToken(travisToken)
	// TODO: Or use github authentication (note: UsingGithubToken() calls UsingTravisToken() eventually anyway)
	// gitHubToken = os.Getenv("TRAVIS_GITHUB_PERSONAL_ACCESS_TOKEN")
	// travisToken, resp, err := client.Authentication.UsingGithubToken(ctx, gitHubToken)
	r.Nil(err)

	name := "TEST"
	sink := sink.TravisCiSink{
		RepoSlug: repoSlug,
		Client:   client,
	}
	ctx := context.Background()

	// create key
	body := travis.EnvVarBody{Name: name, Value: "", Public: false}
	e, _, err := sink.Client.EnvVars.CreateByRepoSlug(ctx, repoSlug, &body)
	r.Nil(err)

	// rotate key
	creds, err := (&source.DummySource{}).Read()
	r.Nil(err)
	err = sink.Write(ctx, name, creds[source.Secret])
	r.Nil(err)

	// delete key
	_, err = sink.Client.EnvVars.DeleteByRepoSlug(ctx, repoSlug, *e.Id)
	r.Nil(err)
}
