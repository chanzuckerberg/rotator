package sink

// Creates a new keypair & connects to snowflake? https://www.systutorials.com/how-to-generate-rsa-private-and-public-key-pair-in-go-lang/
// Figure out which packages to use from this list: https://github.com/chanzuckerberg/terraform-provider-snowflake/blob/main/go.mod
// Figure out what API tokens are needed? Ugh....
// Write tests to make sure that things are behaving okay

type SnowflakeSink interface {
	BaseSink	`yaml:",inline"`
	// Snowflake connection details (like token, etc.)
	user 		string `yaml:"user"`
}


func (sink *SnowflakeSink) Write(ctx context.Context, user string) error {
	// Generate a new set of public and private keys and run snowflake query
	
}