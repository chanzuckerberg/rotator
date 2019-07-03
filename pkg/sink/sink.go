package sink

// Sink is the interface for all credential sinks.
//
// Write writes each key value pair in creds to the underlying
// sink. It returns any error encountered that caused the write
// to stop early.
//
// Kind returns the kind of sink.
type Sink interface {
	Write(creds map[string]string) error
	Kind() Kind
}

type Kind string

type Error string

func (e Error) Error() string { return string(e) }

const (
	KindBuf        Kind  = "buf"
	KindTravisCi   Kind  = "travisCi"
	ErrUnknownKind Error = "unknown sink"
)

type Sinks []Sink

func (sinks Sinks) MarshalYAML() (interface{}, error) {
	var yamlSinks []map[string]string
	for _, s := range sinks {
		switch s.Kind() {
		case KindBuf:
			yamlSinks = append(yamlSinks, map[string]string{"kind": string(KindBuf)})
		case KindTravisCi:
			sink := s.(*TravisCiSink)
			yamlSinks = append(yamlSinks,
				map[string]string{
					"kind":      string(KindTravisCi),
					"repo_slug": sink.RepoSlug,
				})
		default:
			return nil, ErrUnknownKind
		}
	}
	return yamlSinks, nil
}
