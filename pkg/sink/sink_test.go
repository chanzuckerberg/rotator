package sink

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var sink_types = []Kind{
	KindBuf,
	KindTravisCi,
	KindCircleCi,
	KindGithubActionsSecret,
	KindAwsParamStore,
	KindAwsSecretsManager,
	KindStdout,
	KindHeroku,
}

func TestMarshalSinks(t *testing.T) {
	r := require.New(t)

	// initialize all sinks
	allSinks := Sinks{
		NewBufSink(),
		NewTravisCiSink(),
		NewCircleCiSink(),
		NewGitHubActionsSecretSink(),
		NewAwsParamSink(),
		NewAwsSecretsManagerSink(),
		NewStdoutSink(),
		NewHerokuSink(),
	}
	yamlSinks, err := allSinks.MarshalYAML()
	r.NoError(err)
	yamlSinksList, ok := yamlSinks.([]map[string]interface{})
	r.True(ok)
	r.Equal(len(sink_types), len(yamlSinksList))
}
