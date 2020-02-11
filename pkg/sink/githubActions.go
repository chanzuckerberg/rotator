package sink

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
	"golang.org/x/crypto/nacl/box"
)

const (
	githubAPIRetryAttempts = 5
	githubAPIRetrySleep    = time.Second
)

// GithubActionsSink returns the
type GithubActionsSink struct {
	BaseSink `yaml:",inline"`

	Owner  string         `yaml:"owner"`
	Repo   string         `yaml:"repo"`
	Client *github.Client `yaml:"client"`
}

// WithGithubClient configures a github client for this sink
func (sink *GithubActionsSink) WithGithubClient(client *github.Client) *GithubActionsSink {
	sink.Client = client
	return sink
}

// Write updates the value of the env var with the specified name
// for the given repo.
func (sink *GithubActionsSink) Write(ctx context.Context, name string, val string) error {
	// get the repo public key
	publicKey, resp, err := sink.Client.Actions.GetPublicKey(ctx, sink.Owner, sink.Repo)
	if err != nil {
		return errors.Wrapf(err, "unable to get public key in Github for repo %s/%s", sink.Owner, sink.Repo)
	}
	if resp.StatusCode < 200 || 300 <= resp.StatusCode {
		return errors.New(fmt.Sprintf("unable to get public key in Github for repo %s/%s: invalid http status: %s", sink.Owner, sink.Repo, resp.Status))
	}

	decodedPublicKey, err := base64.StdEncoding.DecodeString(*publicKey.Key)
	if err != nil {
		return errors.Wrapf(err, "unable to base64 decode public key from Github for repo %s/%s", sink.Owner, sink.Repo)
	}
	if len(decodedPublicKey) != 32 {
		return errors.New(fmt.Sprintf("public key in Github for repo %s/%s of length %d is not expected length 32", sink.Owner, sink.Repo, len(decodedPublicKey)))
	}
	var fixedSizePublicKey [32]byte
	copy(fixedSizePublicKey[:], decodedPublicKey)
	encrypted, err := box.SealAnonymous(nil, []byte(val), &fixedSizePublicKey, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to encrypt Github secret %s for repo %s/%s", name, sink.Owner, sink.Repo)
	}

	eSecret := &github.EncryptedSecret{
		Name:           name,
		KeyID:          *publicKey.KeyID,
		EncryptedValue: base64.StdEncoding.EncodeToString(encrypted),
	}

	return sink.createOrUpdate(ctx, eSecret)
}

func (sink *GithubActionsSink) createOrUpdate(ctx context.Context, eSecret *github.EncryptedSecret) error {
	f := func(ctx context.Context) error {
		resp, err := sink.Client.Actions.CreateOrUpdateSecret(ctx, sink.Owner, sink.Repo, eSecret)
		if err != nil {
			return errors.Wrapf(err, "unable to create or update env var %s in Github for repo %s/%s", eSecret.Name, sink.Owner, sink.Repo)
		}
		if resp.StatusCode < 200 || 300 <= resp.StatusCode {
			return errors.New(fmt.Sprintf("unable to create or update env var %s in Github for repo %s/%s: invalid http status: %s", eSecret.Name, sink.Owner, sink.Repo, resp.Status))
		}
		return nil
	}

	// Uses retry function defined in travisCI.go
	return retry(ctx, githubAPIRetryAttempts, githubAPIRetrySleep, f)
}

// Kind returns the kind of this sink
func (sink *GithubActionsSink) Kind() Kind {
	return KindGithubActions
}
