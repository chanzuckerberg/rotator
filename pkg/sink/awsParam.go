package sink

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/pkg/errors"
)

type AwsParamSink struct {
	BaseSink `yaml:",inline"`

	RoleArn    string         `yaml:"role_arn"`
	ExternalID string         `yaml:"external_id"`
	Region     string         `yaml:"region"`
	Client     *cziAws.Client `yaml:"client"`
}

func NewAwsParamSink() *AwsParamSink {
	return &AwsParamSink{}
}

// Write updates the value of the the parameter with the given name in the
// underlying AWS Parameter Store.
func (sink *AwsParamSink) Write(ctx context.Context, name string, val interface{}) error {

	switch writeVal := val.(type) {
	case string:
		svc := sink.Client.SSM.Svc

		// check parameter exists in parameter store
		out, err := svc.GetParameterWithContext(ctx, &ssm.GetParameterInput{
			Name: &name,
		})
		if err != nil {
			return errors.Wrapf(err, "%s: unable to get parameter from aws parameter store", name)
		}

		// update parameter value
		in := &ssm.PutParameterInput{
			Name:      &name,
			Value:     &writeVal,
			Type:      out.Parameter.Type,
			Overwrite: aws.Bool(true),
		}
		_, err = svc.PutParameterWithContext(ctx, in)
		return errors.Wrapf(err, "%s: unable to edit parameter in aws parameter store", name)
	default:
		return errors.Errorf("AWSParam Sink doesn't support writing type %T", writeVal)
	}
}

func (sink *AwsParamSink) Kind() Kind {
	return KindAwsParamStore
}
