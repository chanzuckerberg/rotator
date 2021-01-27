package source

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarshalSources(t *testing.T) {
	r := require.New(t)

	// initialize all sources
	allSources := []Source{
		NewAwsIamSource(),
		NewDummySource(),
		NewEnvSource(),
	}
	for _, source := range allSources {
		output, err := source.MarshalYAML()
		r.NoError(err)
		r.NotNil(output)
	}
}
