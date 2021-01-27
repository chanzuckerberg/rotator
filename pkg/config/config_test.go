package config_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/chanzuckerberg/rotator/pkg/config"
	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestFromFile(t *testing.T) {
	r := require.New(t)

	tmpFile, err := ioutil.TempFile("", "tmpConfig")
	r.Nil(err)
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	c1 := &config.Config{Secrets: []config.Secret{}}
	bytes, err := yaml.Marshal(c1)
	r.Nil(err)
	_, err = tmpFile.Write(bytes)
	r.Nil(err)

	c2, err := config.FromFile(tmpFile.Name())
	r.Nil(err)

	r.Equal(c1, c2)
}

func TestSingleStringPairs(t *testing.T) {
	r := require.New(t)
	tmpFile, err := ioutil.TempFile("", "tmpConfig")
	r.Nil(err)
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())
	testSink := sink.NewStdoutSink()
	testSink.WithKeyToName(map[string]string{
		"TEST_ENV": "test_env",
	})
	testSource := source.NewEnvSource()
	testSource.WithName("blah")

	c1 := &config.Config{Secrets: []config.Secret{
		{
			Name:   "foo",
			Source: testSource,
			Sinks: sink.Sinks{
				testSink,
			},
		},
	}}
	spew.Dump(c1)
	bytes, err := yaml.Marshal(c1)
	r.Nil(err)
	_, err = tmpFile.Write(bytes)
	r.Nil(err)
	c2, err := config.FromFile(tmpFile.Name())
	r.Nil(err)

	r.Equal(c1, c2)
}

// Todo: Think about whether the dummy source suffices
func TestConfigWithLists(t *testing.T) {
	r := require.New(t)
	tmpFile, err := ioutil.TempFile("", "tmpConfig")
	r.Nil(err)
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	c1 := &config.Config{Secrets: []config.Secret{
		// add content here
	}}
	// Marshal (just single key-pair values)
	bytes, err := yaml.Marshal(c1)
	r.Nil(err)
	_, err = tmpFile.Write(bytes)
	r.Nil(err)
	// read file
	c2, err := config.FromFile(tmpFile.Name())
	r.Nil(err)

	r.Equal(c1, c2)
}

func TestConfigWithCustomStructs(t *testing.T) {
	r := require.New(t)
	tmpFile, err := ioutil.TempFile("", "tmpConfig")
	r.Nil(err)
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	c1 := &config.Config{Secrets: []config.Secret{
		// Add content here!!!
	}}
	// Marshal (just single key-pair values)
	bytes, err := yaml.Marshal(c1)
	r.Nil(err)
	_, err = tmpFile.Write(bytes)
	r.Nil(err)
	// read file
	c2, err := config.FromFile(tmpFile.Name())
	r.Nil(err)

	r.Equal(c1, c2)
}
