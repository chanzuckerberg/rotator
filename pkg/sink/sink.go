package sink

import (
	"context"
	"fmt"
)

// Sink is the interface for all credential sinks.
//
// Write updates the value of the credential with the given name
// in the underlying sink.
// Unless otherwise specified, Write overwrites existing
// values and does not create new credentials in the sink.
// It returns any error encountered that caused the write
// to stop early.
//
// GetKeyToName returns the KeyToName field's value.
//
// Kind returns the kind of sink.
type Sink interface {
	Write(ctx context.Context, name string, val interface{}) error //TODO(aku): Might be worth making val an interface
	GetKeyToName() map[string]string
	Kind() Kind
}

type Kind string

const (
	KindBuf                 Kind = "Buffer"
	KindTravisCi            Kind = "TravisCI"
	KindCircleCi            Kind = "CircleCI"
	KindGithubActionsSecret Kind = "GitHubActionsSecret"
	KindAwsParamStore       Kind = "AWSParameterStore"
	KindAwsSecretsManager   Kind = "AWSSecretsManager"
	KindStdout              Kind = "Stdout"
	KindHeroku              Kind = "Heroku"
)

type Sinks []Sink

func (sinks Sinks) MarshalYAML() (interface{}, error) {
	var yamlSinks []map[string]interface{}
	for _, s := range sinks {
		switch s.Kind() {
		case KindStdout:
			sink := s.(*StdoutSink)
			yamlSinks = append(yamlSinks,
				map[string]interface{}{
					"key_to_name": sink.KeyToName,
					"kind":        string(KindStdout),
				})
		case KindBuf:
			sink := s.(*BufSink)
			yamlSinks = append(yamlSinks,
				map[string]interface{}{
					"key_to_name": sink.KeyToName,
					"kind":        string(KindBuf),
				})
		case KindTravisCi:
			sink := s.(*TravisCiSink)
			yamlSinks = append(yamlSinks,
				map[string]interface{}{
					"key_to_name": sink.KeyToName,
					"kind":        string(KindTravisCi),
					"repo_slug":   sink.RepoSlug,
				})
		case KindAwsParamStore:
			sink := s.(*AwsParamSink)
			yamlSinks = append(yamlSinks,
				map[string]interface{}{
					"kind":        string(KindAwsParamStore),
					"role_arn":    sink.RoleArn,
					"external_id": sink.ExternalID,
					"region":      sink.Region,
				})
		case KindAwsSecretsManager:
			sink := s.(*AwsSecretsManagerSink)
			yamlSinks = append(yamlSinks,
				map[string]interface{}{
					"kind":        string(KindAwsSecretsManager),
					"role_arn":    sink.RoleArn,
					"external_id": sink.ExternalID,
					"region":      sink.Region,
				})
		case KindCircleCi:
			sink := s.(*CircleCiSink)
			yamlSinks = append(yamlSinks,
				map[string]interface{}{
					"kind":        string(KindCircleCi),
					"key_to_name": sink.KeyToName,
					"account":     sink.Account,
					"repo":        sink.Repo,
				})
		case KindGithubActionsSecret:
			sink := s.(*GitHubActionsSecretSink)
			yamlSinks = append(yamlSinks,
				map[string]interface{}{
					"kind":        string(KindGithubActionsSecret),
					"key_to_name": sink.KeyToName,
					"repo":        sink.repo,
					"owner":       sink.owner,
				})
		case KindHeroku:
			sink := s.(*HerokuSink)
			yamlSinks = append(yamlSinks,
				map[string]interface{}{
					"kind":        string(KindHeroku),
					"key_to_name": sink.KeyToName,
				})
		default:
			return nil, fmt.Errorf("unknown sink kind: %s", s.Kind())
		}
	}
	return yamlSinks, nil
}

func (sink *BaseSink) WithKeyToName(m map[string]string) *BaseSink {
	sink.KeyToName = m
	return sink
}
