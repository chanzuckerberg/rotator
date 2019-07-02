package source

import "math/rand"

// A DummySource represents a source that generates random data.
type DummySource struct{}

// Keys for the map returned by Read()
const (
	Secret string = "secret"
)

// Read returns a random alphanumeric string of length 10.
func (src *DummySource) Read() (map[string]string, error) {
	// reference: https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
	const bytesSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 10)
	for i := range b {
		b[i] = bytesSet[rand.Intn(len(bytesSet))]
	}
	return map[string]string{Secret: string(b)}, nil
}

func (src *DummySource) Kind() Kind {
	return KindDummy
}
