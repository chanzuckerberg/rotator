package source_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/google/go-github/v26/github"
	"github.com/stretchr/testify/require"
)

var (
	userName    = os.Getenv("GH_USERNAME")
	password    = os.Getenv("GH_PASSWORD")
	otp         = os.Getenv("GH_OTP")
	note        = "test"
	fingerprint = time.Now().String()
)

func TestReadFromGitHub(t *testing.T) {
	r := require.New(t)

	// set up github client
	bat := github.BasicAuthTransport{
		Username: userName,
		Password: password,
		OTP:      otp,
	}
	tc := bat.Client()
	client := github.NewClient(tc)

	// create test token to rotate
	req := &github.AuthorizationRequest{
		Scopes:      []github.Scope{github.ScopeRepo},
		Note:        &note,
		Fingerprint: &fingerprint,
	}
	auth, _, err := client.Authorizations.Create(context.Background(), req)
	r.Nil(err)

	src := source.GitHubSource{
		UserName: userName,
		Client:   client,
		TokenID:  auth.GetID(),
		MaxAge:   time.Nanosecond,
	}
	creds, err := src.Read()
	r.Nil(err)
	r.NotNil(creds)
	r.NotEqual(auth.GetID(), creds[source.GHTokenID])
}
