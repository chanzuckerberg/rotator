package config

import (
	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Version int      `yaml:"version"`
	Secrets []Secret `yaml:"secrets"`
}

type Secret struct {
	Name   string        `yaml:"name"`
	Source source.Source `yaml:"source"`
	Sinks  sink.Sinks    `yaml:"sinks"`
}

func (secret *Secret) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var secretFields map[string]interface{}
	if err := unmarshal(&secretFields); err != nil {
		return errors.Wrap(err, "incorrect secret config")
	}
	logrus.Debug(secretFields)

	// unmarshal secret.Name
	name, ok := secretFields["name"]
	if !ok {
		return errors.New("missing name in secret config")
	}
	secret.Name, ok = name.(string)
	if !ok {
		return errors.New("incorrect name format in secret config")
	}

	// unmarshal secret.Source
	src, ok := secretFields["source"]
	if !ok {
		return errors.New("missing source in secret config")
	}
	// convert map[interface {}]interface {} to map[string]string
	srcMap, ok := src.(map[interface{}]interface{})
	if !ok {
		return errors.New("incorrect source format in secret config")
	}
	srcMapStr := make(map[string]string)
	for k, v := range srcMap {
		strK, ok := k.(string)
		if !ok {
			return errors.New("incorrect source format in secret config")
		}
		strV, ok := v.(string)
		if !ok {
			return errors.New("incorrect source format in secret config")
		}
		srcMapStr[strK] = strV
	}
	// determine source kind
	srcKind, ok := srcMapStr["kind"]
	if !ok {
		return errors.New("missing kind in source config")
	}
	switch source.Kind(srcKind) {
	case source.KindDummy:
		secret.Source = &source.DummySource{}
	case source.KindAws:
		// set up AWS session and IAM service client
		sess, err := session.NewSession(&aws.Config{})
		if err != nil {
			return errors.Wrap(err, "unable to set up aws session: make sure you have a shared credentials file or your environment variables set")
		}
		roleArn, ok := srcMapStr["role_arn"]
		if !ok {
			return errors.New("missing role_arn in aws iam source config")
		}
		sess.Config.Credentials = stscreds.NewCredentials(sess, roleArn) // the new Credentials object wraps the AssumeRoleProvider
		awsClient := cziAws.New(sess).WithIAM(sess.Config)

		// parse max age
		maxAgeStr, ok := srcMapStr["max_age"]
		if !ok {
			return errors.New("missing max_age in aws iam source config")
		}
		maxAge, err := time.ParseDuration(maxAgeStr)
		if err != nil {
			return errors.Wrap(err, "incorrect max_age format in aws iam source config")
		}
		secret.Source = source.NewAwsIamSource().WithUserName(srcMapStr["username"]).WithAwsClient(awsClient).WithMaxAge(maxAge)
	default:
		return source.ErrUnknownKind
	}

	// unmarshall sinks
	sinks, ok := secretFields["sinks"]
	if !ok {
		return errors.New("missing sinks in secret config")
	}
	is, ok := sinks.([]interface{})
	if !ok {
		return errors.New("incorrect sinks format in secret config")
	}
	for _, i := range is {
		// convert map[interface {}]interface {} to map[string]string
		sinkMap, ok := i.(map[interface{}]interface{})
		if !ok {
			return errors.New("incorrect sink format in secret config")
		}
		sinkMapStr := make(map[string]string)
		for k, v := range sinkMap {
			strK, ok := k.(string)
			if !ok {
				return errors.New("incorrect sink format in secret config")
			}
			strV, ok := v.(string)
			if !ok {
				return errors.New("incorrect sink format in secret config")
			}
			sinkMapStr[strK] = strV
		}
		// determine sink kind
		sinkKind, ok := sinkMapStr["kind"]
		if !ok {
			return errors.New("missing kind in source config")
		}
		switch sink.Kind(sinkKind) {
		case sink.KindBuf:
			secret.Sinks = append(secret.Sinks, sink.NewBufSink())
		default:
			return sink.ErrUnknownKind
		}
	}

	return nil
}

func (secret Secret) MarshalYAML() (interface{}, error) {
	secretFields := make(map[string]interface{})

	// marshall secret.Name
	secretFields["name"] = secret.Name

	// marshall secret.Source
	switch secret.Source.Kind() {
	case source.KindDummy:
		secretFields["source"] = map[string]string{"kind": string(source.KindDummy)}
	case source.KindAws:
		awsIamSrc := secret.Source.(*source.AwsIamSource)
		clientBytes, err := yaml.Marshal(awsIamSrc.Client)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal AWS client")
		}
		secretFields["source"] = map[string]string{"kind": string(source.KindAws),
			"username": awsIamSrc.UserName,
			"role_arn": awsIamSrc.RoleArn,
			"client":   string(clientBytes),
			"max_age":  awsIamSrc.MaxAge.String(),
		}
	default:
		return nil, errors.New("Unrecognized source")
	}

	// marshall secret.Sinks
	secretFields["sinks"] = secret.Sinks
	return &secretFields, nil
}

func FromFile(file string) (*Config, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read config %s", file)
	}

	conf := &Config{}
	err = yaml.Unmarshal(b, conf)
	return conf, errors.Wrap(err, "Could not unmarshal config")
}
