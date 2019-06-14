package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Version int      `yaml:"version"`
	Secrets []Secret `yaml:"secrets"`
}

type Secret struct {
	Name         string        `yaml:"name"`
	Source       Source        `yaml:"source"`
	Destinations []Destination `yaml:"destinations"`
}

type Source struct {
	Kind string `yaml:"kind"`
}

type Destination struct {
	Kind string `yaml:"kind"`
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
