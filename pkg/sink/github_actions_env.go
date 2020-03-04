package sink

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/google/go-github/v29/github"
	"github.com/pkg/errors"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/oauth2"
)

const (
	gitHubPubKeyLen = 32
)

type GitHubActionsEnvSink struct {
	BaseSink `yaml:",inline"`

	owner string // github organization owner
	repo  string // github repo

	client *github.Client
}

func (s *GitHubActionsEnvSink) WithStaticTokenAuthClient(token string, owner string, repo string) *GitHubActionsEnvSink {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	return s.WithClient(client, owner, repo)
}

func (s *GitHubActionsEnvSink) WithClient(client *github.Client, owner string, repo string) *GitHubActionsEnvSink {
	s.client = client
	s.owner = owner
	s.repo = repo

	return s
}

func (s *GitHubActionsEnvSink) Write(ctx context.Context, name string, value string) error {
	f := func(ctx context.Context) error {

		receiverPublicKey, _, err := s.client.Actions.GetPublicKey(ctx, s.owner, s.repo)
		if err != nil {
			return errors.Wrapf(err, "could not fetch %s/%s public key", s.owner, s.repo)
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
			rand.Reader,
		)
		if err != nil {
			return errors.Wrap(err, "error encrypted github secret")
		}

		encryptedSecret := &github.EncryptedSecret{
			Name:           name,
			KeyID:          receiverPublicKey.GetKeyID(),
			EncryptedValue: base64.StdEncoding.EncodeToString(out),
		}

		_, err = s.client.Actions.CreateOrUpdateSecret(
			ctx,
			s.owner,
			s.repo,
			encryptedSecret,
		)
		return errors.Wrap(err, "could not write encrypted secret to GitHub")
	}

	return retry(ctx, defaultRetryAttempts, defaultRetrySleep, f)
}

func (s *GitHubActionsEnvSink) Kind() Kind {
	return KindGithubActionsEnv
}
