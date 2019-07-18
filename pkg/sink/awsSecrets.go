package sink

import (
	"context"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
)

type AwsSecretsManagerSink struct {
	RoleArn string         `yaml:"role_arn"`
	Region  string         `yaml:"region"`
	Client  *cziAws.Client `yaml:"client"`
}

func (sink *AwsSecretsManagerSink) Write(creds map[string]string) error {
	var errs *multierror.Error
	ctx := context.Background()
	svc := sink.Client.SecretsManager.Svc

	for name, val := range creds {
		// update secret value
		in := &secretsmanager.PutSecretValueInput{
			SecretId:     &name,
			SecretString: &val,
		}
		_, err := svc.PutSecretValueWithContext(ctx, in)
		if err != nil {
			logrus.Errorf("%s: unable to store a new encrypted secret value in aws secrets manager: %s", name, err)
			errs = multierror.Append(errs, err)
			continue
		}
	}
	return errs.ErrorOrNil()
}

func (sink *AwsSecretsManagerSink) Kind() Kind {
	return KindAwsSecretsManager
}
