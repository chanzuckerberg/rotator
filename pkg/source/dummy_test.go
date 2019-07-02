package source_test

import (
	"testing"

	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/stretchr/testify/require"
)

func TestReadFromDummySource(t *testing.T) {
	r := require.New(t)

	src := source.DummySource{}
	creds, err := src.Read()
	r.Nil(err)
	r.Len(creds[source.Secret], 10)
}
