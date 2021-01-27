package source

import (
	"fmt"
	"os"
)

// Env is a source that reads an environment variable
type Env struct {
	Name string `yaml:"name"`
}

// NewEnvSource returns a new env soruce
func NewEnvSource() *Env {
	return &Env{}
}

// Kind returns the kind of this source
func (e *Env) Kind() Kind {
	return KindEnv
}

// WithName sets Env's target environment variable
func (e *Env) WithName(name string) *Env {
	e.Name = name
	return e
}

func (e *Env) Read() (map[string]interface{}, error) {
	env, present := os.LookupEnv(e.Name)

	if !present {
		return nil, fmt.Errorf("Environment variable %s not present", e.Name)
	}
	envPair := make(map[string]interface{})
	envPair[e.Name] = env
	return envPair, nil
}
