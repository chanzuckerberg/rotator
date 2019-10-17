package source_test

import (
	"os"
	"testing"

	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestReadFromEnvSource(t *testing.T) {
	r := require.New(t)
	envName := uuid.New()

	src := source.NewEnvSource().WithName(envName.String())
	r.Equal(source.KindEnv, src.Kind())

	// Set it and then forget it
	r.NoError(os.Setenv(envName.String(), "testo"))
	defer os.Unsetenv(envName.String())

	vals, err := src.Read()
	r.NoError(err)
	r.Equal("testo", vals[envName.String()])

	// Error if not present
	envNotPresent := uuid.New()
	srcNotPresent := source.NewEnvSource().WithName(envNotPresent.String())

	vals, err = srcNotPresent.Read()
	r.Error(err, "Environment variable")
}
