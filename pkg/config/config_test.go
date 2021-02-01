package config_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/chanzuckerberg/rotator/cmd"
	"github.com/chanzuckerberg/rotator/pkg/config"
	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/chanzuckerberg/rotator/pkg/source"
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

	c1 := &config.Config{
		Version: 1,
		Secrets: []config.Secret{
			{
				Name:   "foo",
				Source: testSource,
				Sinks: sink.Sinks{
					testSink,
				},
			},
		},
	}
	bytes, err := yaml.Marshal(c1)
	r.Nil(err)
	r.NotNil(bytes)
	_, err = tmpFile.Write(bytes)
	r.NoError(err)
	c2, err := config.FromFile(tmpFile.Name())
	r.NoError(err)

	r.Equal(c1, c2)
}

// A DummySource represents a source that generates random data.
type ListSource struct {
	BulletPoints []string
}

// Define Read function
func (listSrc *ListSource) Read() (map[string]interface{}, error) {
	dummyIface := make(map[string]interface{})
	dummyIface[source.Secret] = []string{"item1", "item2", "item3"}
	return dummyIface, nil
}

// Kind returns the kind of this source
func (listSrc *ListSource) Kind() source.Kind {
	return source.Kind("listSource")
}

// A DummySource represents a source that generates random data.
type ListSink struct {
	Secret []string
}

func (listSink *ListSink) Write(ctx context.Context, name string, val interface{}) error {
	// enforce that interface{} is of type []string
	fmt.Printf("sink:stdout: \n name: %s, val: %#v\n", name, val)
	return nil
}

func (listSink *ListSink) GetKeyToName() map[string]string {
	return map[string]string{
		"dummyVal": "This won't be used",
	}
}

func NewListSink() *ListSink {
	return &ListSink{}
}

func (*ListSink) Kind() sink.Kind {
	return "ListSink"
}

// Todo: Think about whether the dummy source suffices
func TestConfigWithLists(t *testing.T) {
	r := require.New(t)
	tmpFile, err := ioutil.TempFile("", "tmpConfig")
	r.Nil(err)
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())
	sink1 := NewListSink()

	c1 := &config.Config{Secrets: []config.Secret{
		{
			Name:   "listTest",
			Source: &ListSource{},
			Sinks:  []sink.Sink{sink1},
		},
	}}
	// See if rotating works
	err = cmd.RotateSecrets(c1)
	r.NoError(err)
}

// func TestConfigWithCustomStructs(t *testing.T) {
// 	r := require.New(t)
// 	tmpFile, err := ioutil.TempFile("", "tmpConfig")
// 	r.Nil(err)
// 	defer tmpFile.Close()
// 	defer os.Remove(tmpFile.Name())

// 	c1 := &config.Config{Secrets: []config.Secret{
// 		// Add content here!!!
// 	}}
// 	// Marshal (just single key-pair values)
// 	bytes, err := yaml.Marshal(c1)
// 	r.Nil(err)
// 	_, err = tmpFile.Write(bytes)
// 	r.Nil(err)
// 	// read file
// 	c2, err := config.FromFile(tmpFile.Name())
// 	r.Nil(err)

// 	r.Equal(c1, c2)
// }
