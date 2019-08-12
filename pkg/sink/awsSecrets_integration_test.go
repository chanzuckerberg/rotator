// +build integration

package sink_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/stretchr/testify/require"
)

func TestWriteToAwsSecretsManagerSink_Integration(t *testing.T) {
	r := require.New(t)

	// create a Secrets Manager client from a session.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(defaultRegion), // Sessions Manager functions require region configuration
	})
	r.Nil(err)
	sess.Config.Credentials = stscreds.NewCredentials(sess, roleArn) // the new Credentials object wraps the AssumeRoleProvider
	client := cziAws.New(sess).WithSecretsManager(sess.Config)

	sink := sink.AwsSecretsManagerSink{Client: client}
	svc := client.SecretsManager.Svc
	ctx := context.Background()

	// get the secret
	in := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}
	old, err := svc.GetSecretValueWithContext(ctx, in)
	r.Nil(err)

	// rotate the secret
	creds, err := (&source.DummySource{}).Read()
	r.Nil(err)
	err = sink.Write(ctx, secretName, creds[source.Secret])
	r.Nil(err)
	new, err := svc.GetSecretValueWithContext(ctx, in)
	r.Nil(err)

	// check new secret value and other attributes
	r.Equal(creds[source.Secret], *new.SecretString)
	r.NotEqual(*old.VersionId, *new.VersionId)
	r.Equal(*old.ARN, *new.ARN)
}
