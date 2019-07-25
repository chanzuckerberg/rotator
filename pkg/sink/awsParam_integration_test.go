// +build integration

package sink_test

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/stretchr/testify/require"
)

const (
	region = "us-west-2"
)

func TestWriteToAwsParamSink_Integration(t *testing.T) {
	r := require.New(t)

	// Create a SSM client from a session.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region), // SSM functions require region configuration
	})
	r.Nil(err)
	roleArn := os.Getenv("ROLE_ARN")
	sess.Config.Credentials = stscreds.NewCredentials(sess, roleArn) // the new Credentials object wraps the AssumeRoleProvider
	client := cziAws.New(sess).WithSSM(sess.Config)

	sink := sink.AwsParamSink{Client: client}
	svc := client.SSM.Svc
	ctx := context.Background()

	// get the secret
	in := &ssm.GetParameterInput{
		Name:           aws.String(parName),
		WithDecryption: aws.Bool(true),
	}
	old, err := svc.GetParameterWithContext(ctx, in)
	r.Nil(err)

	// rotate the secret
	creds, err := (&source.DummySource{}).Read()
	r.Nil(err)
	err = sink.Write(ctx, parName, creds[source.Secret])
	r.Nil(err)
	new, err := svc.GetParameterWithContext(ctx, in)
	r.Nil(err)

	// check new parameter value and other attributes
	r.Equal(creds[source.Secret], *new.Parameter.Value)
	r.NotEqual(*old.Parameter.Version, *new.Parameter.Version)
	r.Equal(*old.Parameter.Type, *new.Parameter.Type)
	r.Equal(*old.Parameter.ARN, *new.Parameter.ARN)
}
