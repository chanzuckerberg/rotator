package sink_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/stretchr/testify/require"
)

var (
	secretName     = "secret"
	secretVal      = "value"
	fakeSecretName = "non-existing secret"
)

func (ts *TestSuite) TestWriteToAwsSecretsManagerSink() {
	t := ts.T()
	r := require.New(t)

	// mock PutSecretValueWithContext
	in := &secretsmanager.PutSecretValueInput{
		SecretId:     &secretName,
		SecretString: &secretVal,
	}
	out := &secretsmanager.PutSecretValueOutput{
		Name: &secretName,
	}
	ts.mockSecretsManager.On("PutSecretValueWithContext", in).Return(out, nil)

	// write secret to sink
	ts.sink = &sink.AwsSecretsManagerSink{Client: ts.awsClient}
	err := ts.sink.Write(map[string]string{
		secretName: secretVal,
	})
	r.Nil(err)
}

func (ts *TestSuite) TestWriteToAwsSecretsManagerSinkFakeSecret() {
	t := ts.T()
	r := require.New(t)

	// mock PutSecretValueWithContext
	in := &secretsmanager.PutSecretValueInput{
		SecretId:     &fakeSecretName,
		SecretString: &secretVal,
	}
	out := &secretsmanager.PutSecretValueOutput{}
	errNotFound := awserr.New(secretsmanager.ErrCodeResourceNotFoundException, "", nil)
	ts.mockSecretsManager.On("PutSecretValueWithContext", in).Return(out, errNotFound)

	// write non-existing secret to sink
	ts.sink = &sink.AwsSecretsManagerSink{Client: ts.awsClient}
	err := ts.sink.Write(map[string]string{
		fakeSecretName: secretVal,
	})
	r.NotNil(err)
}

func TestWriteToAwsSecretsManagerSink_Integration(t *testing.T) {
	r := require.New(t)

	// create a Secrets Manager client from a session.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region), // Sessions Manager functions require region configuration
	})
	r.Nil(err)
	sess.Config.Credentials = stscreds.NewCredentials(sess, roleArn) // the new Credentials object wraps the AssumeRoleProvider
	client := cziAws.New(sess).WithSecretsManager(sess.Config)

	sink := sink.AwsSecretsManagerSink{Client: client}
	svc := client.SecretsManager.Svc

	// rotate the secret
	creds, err := (&source.DummySource{}).Read()
	r.Nil(err)
	err = sink.Write(creds)
	r.Nil(err)

	// check secret value
	in := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}
	out, err := svc.GetSecretValueWithContext(context.Background(), in)
	r.Nil(err)
	r.Equal(*out.SecretString, creds[secretName])
}
