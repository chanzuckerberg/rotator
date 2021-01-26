package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/hashicorp/go-multierror"
	heroku "github.com/heroku/heroku-go/v5"
	"github.com/jszwedko/go-circleci"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/shuheiktgw/go-travis"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	envCircleCIAuthToken      = "CIRCLECI_AUTH_TOKEN"
	envTravisCIAuthToken      = "TRAVIS_API_AUTH_TOKEN"
	envGitHubActionsAuthToken = "GITHUB_ACTIONS_AUTH_TOKEN"
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

type HerokuEnv struct {
	Bearer_Token string
}

func loadHerokuEnv() (*HerokuEnv, error) {
	env := &HerokuEnv{}
	err := envconfig.Process("heroku", env)
	return env, errors.Wrap(err, "Unable to load all the heroku environment variables")
}

// parseKeyValueMaps converts an interface to the type map[string]string.
func parseKeyValueMaps(iface interface{}) (keyToName map[string]string, err error) {
	mapIface, ok := iface.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("key_to_name value should be a map")
	}
	keyToName = make(map[string]string)
	for ktmKey, ktmVal := range mapIface {
		ktmKeyStr, ok := ktmKey.(string)
		if !ok {
			return nil, errors.Errorf("key_to_name key should be a string. Got %T", ktmKey)
		}
		ktmValStr, ok := ktmVal.(string)
		if !ok {
			return nil, errors.Errorf("key_to_name val should be a string. Got %T", ktmVal)
		}
		keyToName[ktmKeyStr] = ktmValStr
	}
	return keyToName, nil
}

// parseInputMaps converts an interface to the type map[string]interface.
// It also returns a map[string]string if a "key_to_name" entry
// exists, or nil otherwise.
// It returns any error encountered as a result of iface not being of
// the correct type.
func parseInputMaps(iface interface{}) (inputMap map[string]interface{}, keyToName map[string]string, err error) {
	// first convert to map[interface{}]interface{}
	mapIface, ok := iface.(map[interface{}]interface{})
	if !ok {
		return nil, nil, errors.New("interface is not a map")
	}
	inputMap = make(map[string]interface{})

	for k, v := range mapIface {
		strK, ok := k.(string)
		if !ok {
			return nil, nil, errors.New("key is not a string")
		}
		switch strK {
		// Enforce map[string]string for key_to_name values
		case "key_to_name":
			keyToName, err = parseKeyValueMaps(v)
			if err != nil {
				return nil, nil, errors.Wrap(err, "Unable to parse key_to_name")
			}
		// convert rest of the values to map[string]interface
		default:
			inputMap[strK] = v
		}
	}
	return inputMap, keyToName, nil
}

// unmarshalSource converts an interface to a type that implements
// the source.Source interface.
func unmarshalSource(srcIface interface{}) (source.Source, error) {
	// convert srcIface to the type map[string]string
	srcMap, _, err := parseInputMaps(srcIface)
	if err != nil {
		return nil, errors.Wrap(err, "incorrect source format in secret config")
	}

	// determine source kind
	srcKind, ok := srcMap["kind"].(string)
	if !ok {
		return nil, errors.New("missing kind in source config")
	}

	var src source.Source
	switch source.Kind(srcKind) {
	case source.KindDummy:
		src = &source.DummySource{}
	case source.KindAws:
		if err = validate(srcMap, "role_arn", "max_age"); err != nil {
			return nil, errors.Wrap(err, "missing keys in aws iam source config")
		}

		// set up AWS session and IAM service client
		sess, err := session.NewSession(&aws.Config{})
		if err != nil {
			return nil, errors.Wrap(err, "unable to set up aws session: make sure you have a shared credentials file or your environment variables set")
		}
		// create a Credentials object that wraps the AssumeRoleProvider, passing along the external ID if set
		sess.Config.Credentials = stscreds.NewCredentials(sess, srcMap["role_arn"].(string), func(p *stscreds.AssumeRoleProvider) {
			if externalID, ok := srcMap["external_id"].(string); ok && externalID != "" {
				p.ExternalID = &externalID
			}
		})
		client := cziAws.New(sess).WithIAM(sess.Config)

		// parse max age
		maxAge, err := time.ParseDuration(srcMap["max_age"].(string))
		if err != nil {
			return nil, errors.Wrap(err, "incorrect max_age format in aws iam source config")
		}
		src = source.NewAwsIamSource().WithUserName(srcMap["username"].(string)).WithAwsClient(client).WithMaxAge(maxAge)
	case source.KindEnv:
		if err = validate(srcMap, "name"); err != nil {
			return nil, errors.Wrap(err, "missing keys in env source config")
		}
		src = source.NewEnvSource().WithName(srcMap["name"].(string))
	case source.KindSnowflake:
		// TODO: Figure out what values to check for to validate KindSnowflake
		if err = validate(srcMap, "emails"); err != nil {
			return nil, errors.Wrap(err, "Missing email list in snowflake config")
		}
		src = source.NewSnowflakeSource().WithEmails(srcMap["emails"].(string))
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
		sinkMapStr, keyToName, err := parseInputMaps(i)
		if err != nil {
			return nil, errors.Wrap(err, "incorrect sink format in secret config")
		}

		if keyToName == nil {
			return nil, errors.New("missing key_to_name in sink config")
		}

		// determine sink kind
		sinkKind, ok := sinkMapStr["kind"].(string)
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
			travisToken, present := os.LookupEnv(envTravisCIAuthToken)
			if !present {
				return nil, errors.Errorf("missing env var: %s", envTravisCIAuthToken)
			}
			client := travis.NewClient(sink.TravisBaseURL, travisToken)
			sinks = append(sinks, &sink.TravisCiSink{BaseSink: sink.BaseSink{KeyToName: keyToName}, RepoSlug: sinkMapStr["repo_slug"].(string), Client: client})

		case sink.KindCircleCi:
			if err = validate(sinkMapStr, "account", "repo"); err != nil {
				return nil, errors.Wrap(err, "missing keys in circle CI sink config")
			}
			circleToken, present := os.LookupEnv(envCircleCIAuthToken)
			if !present {
				return nil, errors.Errorf("missing env var: %s", envCircleCIAuthToken)
			}
			client := &circleci.Client{Token: circleToken}
			sink := &sink.CircleCiSink{BaseSink: sink.BaseSink{KeyToName: keyToName}}
			sink.WithCircleClient(client, sinkMapStr["account"].(string), sinkMapStr["repo"].(string))
			sinks = append(sinks, sink)

		case sink.KindGithubActionsSecret:
			if err = validate(sinkMapStr, "owner", "repo"); err != nil {
				return nil, errors.Wrapf(err, "missing keys in %s sink", sink.KindGithubActionsSecret)
			}

			githubToken, present := os.LookupEnv(envGitHubActionsAuthToken)
			if !present {
				return nil, errors.Errorf("missing env var: %s", envGitHubActionsAuthToken)
			}

			sink := &sink.GitHubActionsSecretSink{
				BaseSink: sink.BaseSink{KeyToName: keyToName},
			}
			sink = sink.WithStaticTokenAuthClient(githubToken, sinkMapStr["owner"].(string), sinkMapStr["repo"].(string))

			sinks = append(sinks, sink)

		case sink.KindAwsParamStore:
			if err = validate(sinkMapStr, "role_arn", "region"); err != nil {
				return nil, errors.Wrap(err, "missing keys in aws parameter store sink config")
			}

			// create an AWS SSM client from a session
			sess, err := session.NewSession(&aws.Config{
				Region: aws.String(sinkMapStr["region"].(string)), // SSM functions require region configuration
			})
			if err != nil {
				return nil, errors.Wrap(err, "unable to set up aws session: make sure you have a shared credentials file or your environment variables set")
			}
			sess.Config.Credentials = stscreds.NewCredentials(sess, sinkMapStr["role_arn"].(string), func(p *stscreds.AssumeRoleProvider) {
				if externalID, ok := sinkMapStr["external_id"].(string); ok && externalID != "" {
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
				Region: aws.String(sinkMapStr["region"].(string)), // Secrets Manager functions require region configuration
			})
			if err != nil {
				return nil, errors.Wrap(err, "unable to set up aws session: make sure you have a shared credentials file or your environment variables set")
			}
			sess.Config.Credentials = stscreds.NewCredentials(sess, sinkMapStr["role_arn"].(string), func(p *stscreds.AssumeRoleProvider) {
				if externalID, ok := sinkMapStr["external_id"].(string); ok && externalID != "" {
					p.ExternalID = &externalID
				}
			})
			client := cziAws.New(sess).WithSecretsManager(sess.Config)

			sinks = append(sinks, &sink.AwsSecretsManagerSink{BaseSink: sink.BaseSink{KeyToName: keyToName}, Client: client})
		case sink.KindStdout:
			sinks = append(sinks, &sink.StdoutSink{BaseSink: sink.BaseSink{KeyToName: keyToName}})
		case sink.KindHeroku:
			herokuEnv, err := loadHerokuEnv()
			if err != nil {
				return nil, errors.Wrap(err, "Error loading Heroku Environment Variables")
			}
			if err = validate(sinkMapStr, "AppIdentity"); err != nil {
				return nil, errors.Wrap(err, "missing AppIdentity in Heroku sink config")
			}

			// Set up Heroku service
			headers := http.Header{}
			headers.Set("Accept", "application/vnd.heroku+json; version=3")
			transport := heroku.Transport{
				BearerToken:       herokuEnv.Bearer_Token,
				AdditionalHeaders: headers,
			}
			heroku.DefaultClient.Transport = &transport
			herokuService := heroku.NewService(heroku.DefaultClient)

			// Set up herokuSink
			herokuSink := sink.HerokuSink{
				BaseSink:    sink.BaseSink{KeyToName: keyToName},
				AppIdentity: sinkMapStr["AppIdentity"].(string),
			}
			herokuSink.WithKeyToName(keyToName)
			herokuSink.WithHerokuClient(herokuService)

			// Add heroku sink to sinks
			sinks = append(sinks, &herokuSink)
		case sink.KindDatabricks:
			// load environment variables
			// Set up connection to databricks (and Databricks, run a command?)

			databricksSink := sink.DatabricksSink{
				BaseSink: sink.BaseSink{
					KeyToName: keyToName,
				},
			}

			sinks = append(sinks, &databricksSink)
		default:
			return nil, fmt.Errorf("unknown sink kind: %s", sinkKind)
		}
	}
	return sinks, nil
}

// validate returns an error if any key is not present in m
func validate(m map[string]interface{}, keys ...string) error {
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
	case source.KindEnv:
		envSource := secret.Source.(*source.Env)
		secretFields["source"] = map[string]string{
			"kind": string(source.KindEnv),
			"name": envSource.Name,
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
