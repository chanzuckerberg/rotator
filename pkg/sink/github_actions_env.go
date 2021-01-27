package sink

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/oauth2"
)

const (
	gitHubPubKeyLen = 32
)

// GitHubActionsSecretSink holds the configuration for a Github actions secret
type GitHubActionsSecretSink struct {
	BaseSink `yaml:",inline"`

	owner string `yaml:"owner"` // github organization owner
	repo  string `yaml:"repo"`  // github repo

	client *github.Client `yaml:"client"`
}

func NewGitHubActionsSecretSink() *GitHubActionsSecretSink {
	return &GitHubActionsSecretSink{}
}

// WithStaticTokenAuthClient configures a github client for this sink using an oauth token
func (s *GitHubActionsSecretSink) WithStaticTokenAuthClient(token string, owner string, repo string) *GitHubActionsSecretSink {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	return s.WithClient(client, owner, repo)
}

// WithClient configures a github client for this sink
func (s *GitHubActionsSecretSink) WithClient(client *github.Client, owner string, repo string) *GitHubActionsSecretSink {
	s.client = client
	s.owner = owner
	s.repo = repo

	return s
}

// Write updates the value of the env var with the specified name
// for the given repo.
func (s *GitHubActionsSecretSink) Write(ctx context.Context, name string, value string) error {
	f := func(ctx context.Context) error {

		receiverPublicKey, resp, err := s.client.Actions.GetPublicKey(ctx, s.owner, s.repo)
		if err != nil {
			return errors.Wrapf(err, "could not fetch %s/%s public key", s.owner, s.repo)
		}
		if resp.StatusCode < 200 || 300 <= resp.StatusCode {
			return errors.New(fmt.Sprintf("unable to get public key in Github for repo %s/%s: invalid http status: %s", s.owner, s.repo, resp.Status))
		}

		if receiverPublicKey.Key == nil || receiverPublicKey.KeyID == nil {
			return errors.Wrap(err, "invalid GitHub response; receiver key id or public key nil")
		}

		pubKeyByteSlice, err := base64.StdEncoding.DecodeString(receiverPublicKey.GetKey())
		if err != nil {
			return errors.Wrap(err, "could not b64 decode receiver pub key")
		}

		if len(pubKeyByteSlice) != gitHubPubKeyLen {
			return errors.New("receiver public key is not 32 bytes long")
		}

		pubKeyBytes := [gitHubPubKeyLen]byte{}
		copy(pubKeyBytes[:], pubKeyByteSlice[:gitHubPubKeyLen])

		out := []byte{}
		out, err = box.SealAnonymous(
			out,
			[]byte(value),
			&pubKeyBytes,
			nil,
		)
		if err != nil {
			return errors.Wrap(err, "error encrypted github secret")
		}

		encryptedSecret := &github.EncryptedSecret{
			Name:           name,
			KeyID:          receiverPublicKey.GetKeyID(),
			EncryptedValue: base64.StdEncoding.EncodeToString(out),
		}

		resp, err = s.client.Actions.CreateOrUpdateSecret(
			ctx,
			s.owner,
			s.repo,
			encryptedSecret,
		)
		if err != nil {
			return errors.Wrap(err, "could not write encrypted secret to GitHub")
		}
		if resp.StatusCode < 200 || 300 <= resp.StatusCode {
			return errors.New(fmt.Sprintf("unable to create or update env var %s in Github for repo %s/%s: invalid http status: %s", encryptedSecret.Name, s.owner, s.repo, resp.Status))
		}

		return nil
	}

	return retry(ctx, defaultRetryAttempts, defaultRetrySleep, f)
}

// Kind returns the kind of this sink
func (s *GitHubActionsSecretSink) Kind() Kind {
	return KindGithubActionsSecret
}
