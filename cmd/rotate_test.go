package cmd_test

import (
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

func TestRotateSecrets(t *testing.T) {
	tests := []struct {
		name   string
		config config.Config
	}{
		{"non-empty config, dummy source, stdout sink",
			config.Config{
				Version: 0,
				Secrets: []config.Secret{
					config.Secret{
						Name:   "test",
						Source: &source.DummySource{},
						Sinks: sink.Sinks{
							sink.NewBufSink(),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cmd.RotateSecrets(&tt.config); err != nil {
				t.Error(err.Error())
			}
		})
	}
}

func TestRotate(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
	}{
		{"non-empty config, dummy source, stdout sink",
			&config.Config{
				Version: 0,
				Secrets: []config.Secret{
					config.Secret{
						Name:   "test",
						Source: &source.DummySource{},
						Sinks: sink.Sinks{
							sink.NewBufSink(),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)

			tmpFile, err := ioutil.TempFile("", "tmpConfig")
			r.Nil(err)
			defer tmpFile.Close()
			defer os.Remove(tmpFile.Name())

			bytes, err := yaml.Marshal(tt.config)
			r.Nil(err)
			_, err = tmpFile.Write(bytes)
			r.Nil(err)

			configFromFile, err := config.FromFile(tmpFile.Name())
			r.Nil(err)
			r.Equal(tt.config, configFromFile)

			err = cmd.RotateSecrets(configFromFile)
			r.Nil(err)
		})
	}
}
