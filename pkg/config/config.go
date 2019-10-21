package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/shuheiktgw/go-travis"
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

// parseIface converts an interface to the type map[string]string.
// It also returns a second map[string]string if a "key_to_name" entry
// exists, or nil otherwise.
// It returns any error encountered as a result of iface not being of
// the correct type.
func parseIface(iface interface{}) (mapStr map[string]string, keyToName map[string]string, err error) {
	// first convert to map[interface{}]interface{}
	mapIface, ok := iface.(map[interface{}]interface{})
	if !ok {
		return nil, nil, errors.New("interface is not a map")
	}

	// then convert to map[string]string
	mapStr = make(map[string]string)
	for k, v := range mapIface {
		strK, ok := k.(string)
		if !ok {
			return nil, nil, errors.New("key is not a string")
		}
		switch strK {
		case "key_to_name":
			var m map[string]string
			var err error
			keyToName, m, err = parseIface(v)
			if err != nil {
				return nil, nil, errors.Wrap(err, "incorrect key_to_name format")
			}
			if m != nil {
				return nil, nil, errors.New("incorrect key_to_name format")
			}
		default:
			strV, ok := v.(string)
			if !ok {
				return nil, nil, errors.New("value is not a string")
			}
			mapStr[strK] = strV
		}
	}
	return mapStr, keyToName, nil
}

// unmarshalSource converts an interface to a type that implements
// the source.Source interface.
func unmarshalSource(srcIface interface{}) (source.Source, error) {
	// convert srcIface to the type map[string]string
	srcMapStr, _, err := parseIface(srcIface)
	if err != nil {
		return nil, errors.Wrap(err, "incorrect source format in secret config")
	}

	// determine source kind
	srcKind, ok := srcMapStr["kind"]
	if !ok {
		return nil, errors.New("missing kind in source config")
	}
	var src source.Source
	switch source.Kind(srcKind) {
	case source.KindDummy:
		src = &source.DummySource{}
	case source.KindAws:
		if err = validate(srcMapStr, "role_arn", "max_age"); err != nil {
			return nil, errors.Wrap(err, "missing keys in aws iam source config")
		}

		// set up AWS session and IAM service client
		sess, err := session.NewSession(&aws.Config{})
		if err != nil {
			return nil, errors.Wrap(err, "unable to set up aws session: make sure you have a shared credentials file or your environment variables set")
		}
		// create a Credentials object that wraps the AssumeRoleProvider, passing along the external ID if set
		sess.Config.Credentials = stscreds.NewCredentials(sess, srcMapStr["role_arn"], func(p *stscreds.AssumeRoleProvider) {
			if externalID, ok := srcMapStr["external_id"]; ok && externalID != "" {
				p.ExternalID = &externalID
			}
		})
		client := cziAws.New(sess).WithIAM(sess.Config)

		// parse max age
		maxAge, err := time.ParseDuration(srcMapStr["max_age"])
		if err != nil {
			return nil, errors.Wrap(err, "incorrect max_age format in aws iam source config")
		}
		src = source.NewAwsIamSource().WithUserName(srcMapStr["username"]).WithAwsClient(client).WithMaxAge(maxAge)
	case source.KindEnv:
		if err = validate(srcMapStr, "name"); err != nil {
			return nil, errors.Wrap(err, "missing keys in env source config")
		}
		src = source.NewEnvSource().WithName(srcMapStr["name"])
	default:
		return nil, source.ErrUnknownKind
	}
	return src, nil
}

// unmarshalSource converts an interface to the type sink.Sinks.
func unmarshalSinks(sinksIface interface{}) (sink.Sinks, error) {
	is, ok := sinksIface.([]interface{})
	if !ok {
		return nil, errors.New("incorrect sinks format in secret config")
	}
	var sinks sink.Sinks
	for _, i := range is {
		// convert each interface to the type map[string]string and retrieve keyToName mapping
		sinkMapStr, keyToName, err := parseIface(i)
		if err != nil {
			return nil, errors.Wrap(err, "incorrect sink format in secret config")
		}
		if keyToName == nil {
			return nil, errors.New("missing key_to_name in sink config")
		}

		// determine sink kind
		sinkKind, ok := sinkMapStr["kind"]
		if !ok {
			return nil, errors.New("missing kind in sink config")
		}
		switch sink.Kind(sinkKind) {
		case sink.KindBuf:
			sinks = append(sinks, sink.NewBufSink().WithKeyToName(keyToName))
		case sink.KindTravisCi:
			if err = validate(sinkMapStr, "repo_slug"); err != nil {
				return nil, errors.Wrap(err, "missing keys in travis CI sink config")
			}

			// set up Travis CI API client
			client := travis.NewClient(sink.TravisBaseURL, "")
			travisToken := os.Getenv("TRAVIS_API_AUTH_TOKEN")
			err := client.Authentication.UsingTravisToken(travisToken)
			if err != nil {
				return nil, errors.Wrap(err, "unable to authenticate travis API")
			}

			sinks = append(sinks, &sink.TravisCiSink{BaseSink: sink.BaseSink{KeyToName: keyToName}, RepoSlug: sinkMapStr["repo_slug"], Client: client})
		case sink.KindAwsParamStore:
			if err = validate(sinkMapStr, "role_arn", "region"); err != nil {
				return nil, errors.Wrap(err, "missing keys in aws parameter store sink config")
			}

			// create an AWS SSM client from a session
			sess, err := session.NewSession(&aws.Config{
				Region: aws.String(sinkMapStr["region"]), // SSM functions require region configuration
			})
			if err != nil {
				return nil, errors.Wrap(err, "unable to set up aws session: make sure you have a shared credentials file or your environment variables set")
			}
			sess.Config.Credentials = stscreds.NewCredentials(sess, sinkMapStr["role_arn"], func(p *stscreds.AssumeRoleProvider) {
				if externalID, ok := sinkMapStr["external_id"]; ok && externalID != "" {
					p.ExternalID = &externalID
				}
			})
			client := cziAws.New(sess).WithIAM(sess.Config)

			sinks = append(sinks, &sink.AwsParamSink{BaseSink: sink.BaseSink{KeyToName: keyToName}, Client: client})
		case sink.KindAwsSecretsManager:
			if err = validate(sinkMapStr, "role_arn", "region"); err != nil {
				return nil, errors.Wrap(err, "missing keys in aws secrets manager sink config")
			}

			// create an AWS Secrets Manager client from a session
			sess, err := session.NewSession(&aws.Config{
				Region: aws.String(sinkMapStr["region"]), // Secrets Manager functions require region configuration
			})
			if err != nil {
				return nil, errors.Wrap(err, "unable to set up aws session: make sure you have a shared credentials file or your environment variables set")
			}
			sess.Config.Credentials = stscreds.NewCredentials(sess, sinkMapStr["role_arn"], func(p *stscreds.AssumeRoleProvider) {
				if externalID, ok := sinkMapStr["external_id"]; ok && externalID != "" {
					p.ExternalID = &externalID
				}
			})
			client := cziAws.New(sess).WithSecretsManager(sess.Config)

			sinks = append(sinks, &sink.AwsSecretsManagerSink{BaseSink: sink.BaseSink{KeyToName: keyToName}, Client: client})
		default:
			return nil, sink.ErrUnknownKind
		}
	}
	return sinks, nil
}

// validate returns an error if any key is not present in m
func validate(m map[string]string, keys ...string) error {
	var errs *multierror.Error
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			errs = multierror.Append(errs, errors.New(fmt.Sprintf("missing %s", k)))
		}
	}
	return errors.Wrapf(errs.ErrorOrNil(), "config %#v missing keys", m)
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
	srcIface, ok := secretFields["source"]
	if !ok {
		return errors.New("missing source in secret config")
	}
	src, err := unmarshalSource(srcIface)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal source")
	}
	secret.Source = src

	// unmarshall secret.Sinks
	sinksIface, ok := secretFields["sinks"]
	if !ok {
		return errors.New("missing sinks in secret config")
	}
	sinks, err := unmarshalSinks(sinksIface)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal sinks")
	}
	secret.Sinks = sinks
	return nil
}

func (secret Secret) MarshalYAML() (interface{}, error) {
	secretFields := make(map[string]interface{})

	// marshal secret.Name
	secretFields["name"] = secret.Name

	// marshal secret.Source
	switch secret.Source.Kind() {
	case source.KindDummy:
		secretFields["source"] = map[string]string{"kind": string(source.KindDummy)}
	case source.KindAws:
		awsIamSrc := secret.Source.(*source.AwsIamSource)
		secretFields["source"] = map[string]string{"kind": string(source.KindAws),
			"username":    awsIamSrc.UserName,
			"role_arn":    awsIamSrc.RoleArn,
			"external_id": awsIamSrc.ExternalID,
			"max_age":     awsIamSrc.MaxAge.String(),
		}
	default:
		return nil, errors.New("Unrecognized source")
	}

	// marshal secret.Sinks
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
