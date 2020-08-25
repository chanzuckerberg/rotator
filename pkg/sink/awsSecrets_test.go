package sink_test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

const (
	secretName     = "test-secret"
	secretVal      = "value"
	fakeSecretName = "non-existing secret"
)

func (ts *TestSuite) TestWriteToAwsSecretsManagerSink() {
	t := ts.T()
	r := require.New(t)

	// mock PutSecretValueWithContext
	in := &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(secretName),
		SecretString: aws.String(secretVal),
	}
	out := &secretsmanager.PutSecretValueOutput{
		Name: aws.String(secretName),
	}
	ts.mockSecretsManager.EXPECT().PutSecretValueWithContext(gomock.Any(), gomock.Eq(in)).Return(out, nil)

	// write secret to sink
	ts.sink = &sink.AwsSecretsManagerSink{Client: ts.awsClient}
	err := ts.sink.Write(ts.ctx, secretName, secretVal)
	r.Nil(err)
}

func (ts *TestSuite) TestWriteToAwsSecretsManagerSinkFakeSecret() {
	t := ts.T()
	r := require.New(t)

	// mock PutSecretValueWithContext
	in := &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(fakeSecretName),
		SecretString: aws.String(secretVal),
	}
	out := &secretsmanager.PutSecretValueOutput{}
	errNotFound := awserr.New(secretsmanager.ErrCodeResourceNotFoundException, "", nil)
	ts.mockSecretsManager.EXPECT().PutSecretValueWithContext(gomock.Any(), gomock.Eq(in)).Return(out, errNotFound)

	// write non-existing secret to sink
	ts.sink = &sink.AwsSecretsManagerSink{Client: ts.awsClient}
	err := ts.sink.Write(ts.ctx, fakeSecretName, secretVal)
	r.NotNil(err)
}
