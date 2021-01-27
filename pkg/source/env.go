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

func (e *Env) Read() (map[string]string, error) {
	env, present := os.LookupEnv(e.Name)

	if !present {
		return nil, fmt.Errorf("Environment variable %s not present", e.Name)
	}

	return map[string]string{e.Name: env}, nil
}

func (src *Env) MarshalYAML() (interface{}, error) {
	yamlSource := make(map[string]interface{})
	yamlSource["source"] = map[string]string{
		"kind": string(KindEnv),
		"name": src.Name,
	}
	return yamlSource, nil
}
