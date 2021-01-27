package sink

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/shuheiktgw/go-travis"
)

const (
	// TravisBaseURL is the base url for travisCI
	TravisBaseURL string = travis.ApiComUrl
)

// TravisCiSink is a travisCi sink
type TravisCiSink struct {
	BaseSink `yaml:",inline"`

	RepoSlug string         `yaml:"repo_slug"`
	Client   *travis.Client `yaml:"client"`
}

func NewTravisCiSink() *TravisCiSink {
	return &TravisCiSink{}
}

// WithTravisClient configures a travisCI client for this sink
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
	if resp.StatusCode < 200 || 300 <= resp.StatusCode {
		return errors.New(fmt.Sprintf("unable to list env vars in Travis CI for repo %s: invalid http status: %s", sink.RepoSlug, resp.Status))
	}

	body := &travis.EnvVarBody{Name: name, Value: val}

	es := make(map[string]*travis.EnvVar)
	for _, e := range esList {
		es[*e.Name] = e
	}

	// find env var by name
	e, ok := es[name]
	if !ok {
		return sink.create(ctx, body)
	}
	return sink.update(ctx, body, *e.Id)
}

func (sink *TravisCiSink) create(ctx context.Context, body *travis.EnvVarBody) error {
	f := func(ctx context.Context) error {
		_, resp, err := sink.Client.EnvVars.CreateByRepoSlug(ctx, sink.RepoSlug, body)
		if err != nil {
			return errors.Wrapf(err, "unable to create env var %s in TravisCI repo %s", body.Name, sink.RepoSlug)
		}
		if resp.StatusCode < 200 || 300 <= resp.StatusCode {
			return errors.New(fmt.Sprintf("unable to create env var %s in Travis CI for repo %s: invalid http status: %s", body.Name, sink.RepoSlug, resp.Status))
		}
		return nil
	}
	return retry(ctx, defaultRetryAttempts, defaultRetrySleep, f)
}

func (sink *TravisCiSink) update(ctx context.Context, body *travis.EnvVarBody, envID string) error {
	f := func(ctx context.Context) error {
		_, resp, err := sink.Client.EnvVars.UpdateByRepoSlug(ctx, sink.RepoSlug, envID, body)
		if err != nil {
			return errors.Wrapf(err, "unable to update env var %s in TravisCI repo %s", body.Name, sink.RepoSlug)
		}
		if resp.StatusCode < 200 || 300 <= resp.StatusCode {
			return errors.New(fmt.Sprintf("unable to update env var %s in Travis CI for repo %s: invalid http status: %s", body.Name, sink.RepoSlug, resp.Status))
		}
		return nil
	}
	return retry(ctx, defaultRetryAttempts, defaultRetrySleep, f)
}

// Kind returns the kind of this sink
func (sink *TravisCiSink) Kind() Kind {
	return KindTravisCi
}
