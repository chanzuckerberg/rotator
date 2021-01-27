package source

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/pkg/errors"
)

// A DummySource represents a source that generates random data.
type DummySource struct{}

// Keys for the map returned by Read()
const (
	Secret string = "secret"
)

func NewDummySource() *DummySource {
	return &DummySource{}
}

// Read returns a random number of length 16.
func (src *DummySource) Read() (map[string]interface{}, error) {
	// reference: https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
	b := make([]byte, 10)
	_, err := rand.Read(b)
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate random bytes")
	}
	randomSecret := make(map[string]interface{})
	randomSecret[Secret] = base64.URLEncoding.EncodeToString(b)
	return randomSecret, nil
}

func (src *DummySource) Kind() Kind {
	return KindDummy
}
