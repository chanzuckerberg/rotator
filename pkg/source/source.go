package source

import "math/rand"

type Source interface {
	Read() (string, error)
}

type DummySource struct{}

// Generates a random alphanumeric string of length 10.
func (src DummySource) Read() (string, error) {
	// reference: https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
	const bytesSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 10)
	for i := range b {
		b[i] = bytesSet[rand.Intn(len(bytesSet))]
	}
	return string(b), nil
}
