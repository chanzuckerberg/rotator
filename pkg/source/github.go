package source

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-github/v26/github"
	"github.com/pkg/errors"
)

// Keys for the map returned by Read()
const (
	GHTokenID string = "gitHubTokenID"
	GHToken   string = "GitHubToken"
)

type GitHubSource struct {
	UserName string         `yaml:"username"`
	Client   *github.Client `yaml:"client,omitempty"`
	TokenID  int64          `yaml:"token_id"`
	MaxAge   time.Duration  `yaml:"max_age"`
}

func (src *GitHubSource) Read() (map[string]string, error) {
	ctx := context.Background()
	svc := src.Client.Authorizations

	// get token info
	auth, _, err := svc.Get(ctx, src.TokenID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get github token %d for %s", src.TokenID, src.UserName)
	}
	// TODO: do I also need to check Status: 200 OK?

	// do nothing if TOKEN is within max age
	if time.Since(auth.GetCreatedAt().Time) <= src.MaxAge {
		return nil, nil
	}

	// else create new token
	fingerprint := time.Now().String() // differentiates tokens with the same note attribute
	req := &github.AuthorizationRequest{
		Scopes:      auth.Scopes,
		Note:        auth.Note,
		Fingerprint: &fingerprint,
	}
	newAuth, _, err := svc.Create(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create github token for %s", src.UserName)
	}
	spew.Dump(auth)
	spew.Dump(newAuth)

	// delete existing token
	resp, err := svc.Delete(ctx, src.TokenID)
	if err != nil || resp.StatusCode != http.StatusNoContent {
		return nil, errors.Wrapf(err, "unable to delete github token %d for %s", src.TokenID, src.UserName)
	}

	creds := map[string]string{
		GHTokenID: strconv.FormatInt(*newAuth.ID, 10),
		GHToken:   *newAuth.Token,
	}
	// TODO: Update the github token id to the config file
	return creds, nil
}

func (src *GitHubSource) Kind() Kind {
	return KindGitHub
}
