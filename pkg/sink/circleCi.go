package sink

import (
	"context"

	"github.com/jszwedko/go-circleci"
	"github.com/pkg/errors"
)

// CircleCiSink is a circleci sink
type CircleCiSink struct {
	BaseSink `yaml:",inline"`

	Client  *circleci.Client
	Account string
	Repo    string
}

// WithCircleClient configures a circleci client for this sink
func (sink *CircleCiSink) WithCircleClient(client *circleci.Client, account string, repo string) *CircleCiSink {
	sink.Client = client
	sink.Account = account
	sink.Repo = repo

	return sink
}

// Write writes the value of the env var with the specified name for the given repo
func (sink *CircleCiSink) Write(ctx context.Context, name string, val string) error {
	f := func(ctx context.Context) error {
		_, err := sink.Client.AddEnvVar(sink.Account, sink.Repo, name, val)
		return errors.Wrapf(err, "could not write %s to %s/%s", name, sink.Account, sink.Repo)
	}
	return retry(ctx, defaultRetryAttempts, defaultRetrySleep, f)
}

// Kind returns the kind of this sink
func (sink *CircleCiSink) Kind() Kind {
	return KindCircleCi
}
