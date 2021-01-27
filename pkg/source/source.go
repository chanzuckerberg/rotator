package source

import (
	"time"
)

// Source is the interface for all credential sources.
//
// Read reads the secret from the underlying source.
// It returns the secret and any error encountered
// that caused the read to stop early.
//
// Kind returns the kind of sink.
type Source interface {
	Read() (map[string]string, error)
	Kind() Kind
	MarshalYAML() (interface{}, error)
}

type Kind string

type Error string

func (e Error) Error() string { return string(e) }

const (
	KindDummy Kind = "dummy"
	KindAws   Kind = "aws"
	KindEnv   Kind = "env"
)
const (
	ErrUnknownKind Error = "unknown source"
)

const (
	// Default values
	DefaultMaxAge time.Duration = 100 * time.Minute
)
