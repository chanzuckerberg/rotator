package sink

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ssm"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/sirupsen/logrus"
)

type AwsParamSink struct {
	RoleArn string         `yaml:"role_arn"`
	Region  string         `yaml:"region"`
	Client  *cziAws.Client `yaml:"client"`
}

// Write writes each parameter in creds to the underlying AWS Parameter Store.
func (sink *AwsParamSink) Write(creds map[string]string) error {
	ctx := context.Background()
	svc := sink.Client.SSM.Svc

	for name, val := range creds {
		// check parameter exists in parameter store
		out, err := svc.GetParameterWithContext(ctx, &ssm.GetParameterInput{
			Name: &name,
		})
		if err != nil {
			logrus.Errorf("%s: unable to get parameter from aws parameter store", name)
			continue
		}

		// update parameter value
		in := &ssm.PutParameterInput{
			Name:  &name,
			Value: &val,
			Type:  out.Parameter.Type,
		}
		in = in.SetOverwrite(true)
		_, err = svc.PutParameterWithContext(ctx, in)
		if err != nil {
			logrus.Errorf("%s: unable to edit parameter in aws parameter store", name)
			continue
		}
	}
	return nil
}

func (sink *AwsParamSink) Kind() Kind {
	return KindAwsParam
}
