package source

// Source is the interface for all credential sources.
//
// Read reads the secret from the underlying source.
// It returns the secret and any error encountered
// that caused the read to stop early.
//
// Kind returns the kind of sink.
type Source interface {
	Read() (string, error)
	Kind() Kind
}

type Kind string

type Error string

func (e Error) Error() string { return string(e) }

const (
	KindDummy      Kind  = "dummy"
	ErrUnknownKind Error = "unknown source"
)
