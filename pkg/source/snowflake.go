package source

import (
	"context"

	"github.com/pkg/errors"
)

type SnowflakeSource struct {
	Emails               string `yaml:"emails"`
	Account              string `yaml:"account"`
	User                 string `yaml:"user"`
	Role                 string `yaml:"role"`
	Schema               string `yaml:"schema"`
	Database             string `yaml:"database"`
	Warehouse            string `yaml:"warehouse"`
	Private_key_path     string `yaml:"private_key_path"`
	Private_key_password string `yaml:"private_key_password"`
}

// Decide whether to make this struct readable from other files by capitalizing first character
type keypair struct {
	privateKey *string
	publicKey  *string
}

func NewSnowflakeSource() *SnowflakeSource {
	return &SnowflakeSource{}
}

// TODO: accept list of emails in the future
func (src *SnowflakeSource) WithEmails(emails string) *SnowflakeSource {
	src.Emails = emails
	return src
}

func (src *SnowflakeSource) WithAccount(account string) *SnowflakeSource {
	src.Account = account
	return src
}

func (src *SnowflakeSource) WithUser(user string) *SnowflakeSource {
	src.User = user
	return src
}

func (src *SnowflakeSource) WithRole(role string) *SnowflakeSource {
	src.Role = role
	return src
}

func (src *SnowflakeSource) WithSchema(schema string) *SnowflakeSource {
	src.Schema = schema
	return src
}

func (src *SnowflakeSource) WithDatabase(db string) *SnowflakeSource {
	src.Database = db
	return src
}

func (src *SnowflakeSource) WithWarehouse(warehouse string) *SnowflakeSource {
	src.Warehouse = warehouse
	return src
}

func (src *SnowflakeSource) WithPrivate_key_path(privateKeyPath string) *SnowflakeSource {
	src.Private_key_path = privateKeyPath
	return src
}

func (src *SnowflakeSource) WithPrivate_key_password(privateKeyPassword string) *SnowflakeSource {
	src.Private_key_password = privateKeyPassword
	return src
}

// RotateKeys rotates the Snowflake keys for the user specified in src no matter
func (src *SnowflakeSource) RotateKeys(ctx context.Context) (*keypair, error) {
	// svc := src.Client.IAM.Svc // Figure out what a Snowflake client looks like, what info we need

	// Authenticating into Snowflake
	// https://github.com/chanzuckerberg/czi-di-databricks-utils/blob/b7c3e66d2b25cbd03ee8f8bf887aae61a57c0c1a/python/src/czi_datainfra/snowflake_utils.py#L22

	placeholderPrivate := "foo"
	placeholderPublic := "bar"
	// list a user's access keys
	newKeypair := keypair{
		privateKey: &placeholderPrivate,
		publicKey:  &placeholderPublic,
	}
	// return &newKeypair, errors.New("Snowflake RotateKeys() undefined")
	return &newKeypair, nil
}

func (src *SnowflakeSource) Read() (map[string]string, error) {
	newKeypair, err := src.RotateKeys(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "unable to rotate keys")
	}
	if newKeypair == nil {
		return nil, nil
	}
	creds := map[string]string{
		"email":       src.Emails,
		"public_key":  *newKeypair.publicKey,
		"private_key": *newKeypair.privateKey,
	}
	// return creds, errors.New("Snowflake Read() undefined")
	return creds, nil
}

func (src *SnowflakeSource) Kind() Kind {
	return KindSnowflake
}
