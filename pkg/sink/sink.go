package sink

import "context"

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
	Write(ctx context.Context, name string, val string) error
	GetKeyToName() map[string]string
	Kind() Kind
}

type Kind string

type Error string

func (e Error) Error() string { return string(e) }

const (
	KindBuf               Kind  = "Buffer"
	KindTravisCi          Kind  = "TravisCI"
	KindCircleCi                = "CircleCI"
	KindAwsParamStore     Kind  = "AWSParameterStore"
	KindAwsSecretsManager Kind  = "AWSSecretsManager"
	KindStdout            Kind  = "Stdout"
	KindGithubActions     Kind  = "GithubActions"
	ErrUnknownKind        Error = "UnknownSink"
)

type Sinks []Sink

func (sinks Sinks) MarshalYAML() (interface{}, error) {
	var yamlSinks []map[string]interface{}
	for _, s := range sinks {
		switch s.Kind() {
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
		case KindGithubActions:
			sink := s.(*GithubActionsSink)
			yamlSinks = append(yamlSinks,
				map[string]interface{}{
					"key_to_name": sink.KeyToName,
					"kind":        string(KindGithubActions),
					"repo":        sink.Repo,
				})
		default:
			return nil, ErrUnknownKind
		}
	}
	return yamlSinks, nil
}
