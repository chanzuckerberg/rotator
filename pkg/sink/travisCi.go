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
	RepoSlug string         `yaml:"repo_slug"`
	Client   *travis.Client `yaml:"client"`
}

// Write writes each key value pair in creds to the Travis CI client
// for the given repository slug in the underlying sink.
func (sink *TravisCiSink) Write(creds map[string]string) error {
	// make a map of existing env vars
	ctx := context.Background()
	esList, resp, err := sink.Client.EnvVars.ListByRepoSlug(ctx, sink.RepoSlug)
	if err != nil {
		return errors.Wrapf(err, "unable to list env vars in repo %s", sink.RepoSlug)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("unable to list env vars in repo %s: invalid http status: %s", sink.RepoSlug, resp.Status))
	}
	es := make(map[string]*travis.EnvVar)
	for _, e := range esList {
		es[*e.Name] = e
	}

	// for each secret in creds
	for k, v := range creds {
		if e, ok := es[k]; ok {
			// update any existing env var
			body := travis.EnvVarBody{Name: k, Value: v, Public: *e.Public}
			_, resp, err := sink.Client.EnvVars.UpdateByRepoSlug(ctx, sink.RepoSlug, *e.Id, &body)
			if err != nil {
				return errors.Wrapf(err, "unable to update env var %s in repo %s: %s", k, sink.RepoSlug, err)
			} else if resp.StatusCode != http.StatusOK {
				return errors.New(fmt.Sprintf("unable to update env var %s in repo %s: invalid http status: %s", k, sink.RepoSlug, resp.Status))
			}
		} else {
			// else create new env var
			body := travis.EnvVarBody{Name: k, Value: v, Public: false} // TODO: should the Public bool be configurable?
			_, resp, err := sink.Client.EnvVars.CreateByRepoSlug(ctx, sink.RepoSlug, &body)
			if err != nil {
				return errors.Wrapf(err, "unable to create env var %s in repo %s: %s", k, sink.RepoSlug, err)
			} else if resp.StatusCode != http.StatusCreated {
				return errors.New(fmt.Sprintf("unable to create env var %s in repo %s: invalid http status: %s", k, sink.RepoSlug, resp.Status))
			}
		}
	}
	return nil
}

func (sink *TravisCiSink) Kind() Kind {
	return KindTravisCi
}
