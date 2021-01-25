package sink

import (
	"context"

	"github.com/pkg/errors"
)

// Creates a new keypair & connects to databricks? https://www.systutorials.com/how-to-generate-rsa-private-and-public-key-pair-in-go-lang/
// Figure out which packages to use from this list: https://github.com/chanzuckerberg/terraform-provider-databricks/blob/main/go.mod
// Figure out what API tokens are needed? Ugh....
// Write tests to make sure that things are behaving okay

type DatabricksSink struct {
	BaseSink `yaml:",inline"`
	// Databricks connection details (like token, etc.)
	email string `yaml:"email"`
}

func (sink *DatabricksSink) Write(ctx context.Context, email string, publicKey string) error {
	// Generate a new set of public and private keys and run databricks query
	return errors.New("DatabricksSink Write() still undefined")
}

// Kind returns the kind of this sink
func (sink *DatabricksSink) Kind() Kind {
	return KindDatabricks
}
