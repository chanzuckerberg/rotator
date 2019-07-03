package cmd_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/rotator/cmd"
	"github.com/chanzuckerberg/rotator/pkg/config"
	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/shuheiktgw/go-travis"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var (
	roleArn  = "arn:aws:iam::exampleAccount:role/admin"
	userName = "rotator_test"
	repoSlug = "chanzuckerberg/rotator"
)

func TestRotate(t *testing.T) {
	// set up AWS session and IAM service client
	sess, _ := session.NewSession(&aws.Config{})
	sess.Config.Credentials = stscreds.NewCredentials(sess, roleArn) // the new Credentials object wraps the AssumeRoleProvider
	awsClient := cziAws.New(sess).WithIAM(sess.Config)

	// set up Travis CI API client
	travisClient := travis.NewClient(sink.TravisBaseURL, "")
	travisToken := os.Getenv("TRAVIS_API_AUTH_TOKEN")
	_ = travisClient.Authentication.UsingTravisToken(travisToken)

	tests := []struct {
		name   string
		file   string
		config *config.Config
	}{
		{"non-empty config, dummy source, buffer sink",
			"testdata/dummyToBuf.yml",
			&config.Config{
				Version: 0,
				Secrets: []config.Secret{
					config.Secret{
						Name:   "test",
						Source: &source.DummySource{},
						Sinks: sink.Sinks{
							sink.NewBufSink(),
						},
					},
				},
			},
		},
		{"non-empty config, AWS IAM source, travis CI sink",
			"testdata/awsIamtoTravisCi.yml",
			&config.Config{
				Version: 0,
				Secrets: []config.Secret{
					config.Secret{
						Name:   "test",
						Source: source.NewAwsIamSource().WithUserName(userName).WithRoleArn(roleArn).WithAwsClient(awsClient),
						Sinks: sink.Sinks{
							&sink.TravisCiSink{RepoSlug: repoSlug, Client: travisClient},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)

			bytes, err := yaml.Marshal(tt.config)
			r.Nil(err)
			err = ioutil.WriteFile(tt.file, bytes, 0644)
			r.Nil(err)

			configFromFile, err := config.FromFile(tt.file)
			r.Nil(err)
			// r.Equal(tt.config, configFromFile)

			err = cmd.RotateSecrets(configFromFile)
			r.Nil(err)
		})
	}
}
