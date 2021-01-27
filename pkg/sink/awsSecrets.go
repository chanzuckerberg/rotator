package sink

import (
	"context"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/pkg/errors"
)

type AwsSecretsManagerSink struct {
	BaseSink `yaml:",inline"`

	RoleArn    string         `yaml:"role_arn"`
	ExternalID string         `yaml:"external_id"`
	Region     string         `yaml:"region"`
	Client     *cziAws.Client `yaml:"client"`
}

func NewAwsSecretsManagerSink() *AwsSecretsManagerSink {
	return &AwsSecretsManagerSink{}
}

func (sink *AwsSecretsManagerSink) Write(ctx context.Context, name string, val string) error {
	svc := sink.Client.SecretsManager.Svc

	// update secret value
	in := &secretsmanager.PutSecretValueInput{
		SecretId:     &name,
		SecretString: &val,
	}
	_, err := svc.PutSecretValueWithContext(ctx, in)
	return errors.Wrapf(err, "%s: unable to store a new encrypted secret value in aws secrets manager", name)
}

func (sink *AwsSecretsManagerSink) Kind() Kind {
	return KindAwsSecretsManager
}
