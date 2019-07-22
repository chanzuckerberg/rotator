package sink

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/shuheiktgw/go-travis"
)

const (
	// Default values
	TravisBaseURL string = travis.ApiComUrl
)

type TravisCiSink struct {
	BaseSink `yaml:",inline"`

	RepoSlug string         `yaml:"repo_slug"`
	Client   *travis.Client `yaml:"client"`
}

func (sink *TravisCiSink) WithTravisClient(client *travis.Client) *TravisCiSink {
	sink.Client = client
	return sink
}

// Write updates the value of the env var with the specified name
// for the given repository slug using the Travis CI client.
func (sink *TravisCiSink) Write(ctx context.Context, name string, val string) error {
	// make a map of existing env vars
	esList, resp, err := sink.Client.EnvVars.ListByRepoSlug(ctx, sink.RepoSlug)
	if err != nil {
		return errors.Wrapf(err, "unable to list env vars in Travis CI for repo %s", sink.RepoSlug)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("unable to list env vars in Travis CI for repo %s: invalid http status: %s", sink.RepoSlug, resp.Status))
	}
	es := make(map[string]*travis.EnvVar)
	for _, e := range esList {
		es[*e.Name] = e
	}

	// find env var by name
	e, ok := es[name]
	if !ok {
		return errors.New(fmt.Sprintf("env var %s does not exist in Travis CI for repo %s", name, sink.RepoSlug))
	}

	// update env var
	body := travis.EnvVarBody{Name: name, Value: val, Public: *e.Public}
	_, resp, err = sink.Client.EnvVars.UpdateByRepoSlug(ctx, sink.RepoSlug, *e.Id, &body)
	if err != nil {
		return errors.Wrapf(err, "unable to update env var %s in Travis CI for repo %s: %s", name, sink.RepoSlug, err)
	} else if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("unable to update env var %s in Travis CI for repo %s: invalid http status: %s", name, sink.RepoSlug, resp.Status))
	}
	return nil
}

func (sink *TravisCiSink) Kind() Kind {
	return KindTravisCi
}
