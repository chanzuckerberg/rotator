# Rotator

Rotator is a tool for rotating credentials on a regular schedule. It works by reading a YAML configuration file with a list of secret. Each secret consists of a source from which rotator will read new credentials, and one or more sinks to write the new credentials to.

Currently, rotator supports the following sources...
* AWS IAM 

... and sinks:
* Travis CI
* AWS Systems Manager Parameter Store
* AWS Secrets Manager

We go into each source and sink in more detail below.

## Installation

If you have a functional go environment, you can install with:

```bash
$ go get github.com/chanzuckerberg/rotator
```
TODO: set up goreleaser \
TODO: provide other installation options and possibly create a wiki like [chamber](https://github.com/segmentio/chamber/wiki/Installation).

## Usage
Execute rotator with the `rotate` command, passing the `--file/-f` flag to specify the configuration file:
```bash
$ rotator rotate -f config.yaml
```

Below is an example of a configuration file `config.yaml` to rotate credentials for the AWS IAM user `example-user` and write them to the Travis CI repository `example-repo`:
```YAML
version: 1
secrets:
  - name: example_secret
    source:
      kind: aws
      max_age: 1h40m0s 
      role_arn: arn:aws:iam::123456789101:role/admin
      username: example-user
      external_id: ""
    sinks:
      - kind: TravisCI
        key_to_name:
          accessKeyId: EXAMPLE_AWS_ACCESS_KEY_ID
          secretAccessKey: EXAMPLE_AWS_SECRET_ACCESS_KEY   
        repo_slug: example-repo
```

## Flags
`--file/-f`  config file to read from \
`--yes/-y`    assume "yes" to all prompts and run non-interactively

## Sources
All sources must have the following fields in addition to any source-specific fields:

| Name | Description |
|------|-------------|
| kind | The kind of source. Acceptable values: `aws`. |
| max\_age | The max age for a credential before it will be rotated by rotator. The duration string should follow the same format as for [`time.ParseDuration()`](https://golang.org/pkg/time/#ParseDuration) e.g. "2h45m". |

### AWS IAM (`aws`)
| Name | Description | Required |
|------|-------------|:-----:|
| role\_arn | The ARN of the AWS IAM role that rotator should assume. | yes |
| username | The name of the AWS IAM user for rotator to rotate their AWS access keys. | yes |
| external\_id | If set, the [external ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html) is passed to the AWS STS AssumeRole API to assume the IAM Role specified by `role_arn` e.g. if deploying rotator on Kubernetes. | no |

If the `external_id` field is not set, [AWS credentials must be specified](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials) using a shared credentials file or `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.

## Sinks
All sinks must have the following fields in addition to any sink-specific fields:

| Name | Description |
|------|-------------|
| kind | The kind of sink. Acceptable values: `TravisCI`, `AWSParameterStore`, `AWSSecretsManager`. |
| key\_to\_name |  |

### Travis CI
