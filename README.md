# Rotator

Rotator is a tool for rotating credentials on a regular schedule. It works by reading a YAML configuration file with a list of secret. Each secret consists of a source from which rotator will read new credentials, and one or more destinations (from here on referred to as sinks) to write the new credentials to.

Currently, rotator supports the following sources...
* AWS IAM 

... and sinks:
* Travis CI
* AWS Systems Manager Parameter Store
* AWS Secrets Manager

## Table of contents

- [Installation](#installation)
- [Usage](#usage)
    - [Flags](#flags)
- [Monitoring](#monitoring)
- [Sources](#sources)
    - [AWS IAM](#aws-iam-aws)
- [Sinks](#sinks)
    - [Travis CI](#travis-ci-travisci)
    - [AWS Systems Manager Parameter Store](#aws-systems-manager-parameter-store-awsparameterstore)
    - [AWS Secrets Manager](#aws-secrets-manager--awssecretsmanager)
- [Contributing](#contributing)
- [License](#license)

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

### Flags
`-f`, `--file`   config file to read from \
`-y`, `--yes`    assume "yes" to all prompts and run non-interactively

## Monitoring
Configure [Airbrake](https://airbrake.io/) for rotator by setting the `ENV`, `AIRBRAKE_PROJECT_ID`, and `AIRBRAKE_PROJECT_KEY` environment variables.

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

[AWS credentials must be specified](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials) using a shared credentials file or `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.

## Sinks
All sinks must have the following fields in addition to any sink-specific fields:

| Name | Description |
|------|-------------|
| kind | The kind of sink. Acceptable values: `TravisCI`, `AWSParameterStore`, `AWSSecretsManager`. |
| key\_to\_name | A map of source keys to their sink names.* |

> *Rotator parses the credentials from any source as key-value pairs. For example, the credentials for an AWS IAM source will consist of a `AWS_ACCESS_KEY_ID` key and a `AWS_SECRET_ACCESS_KEY` key and their associated values. The `key_to_name` mapping then maps each key to the name of the credential in the sink that rotator should update the value of. This gives users more control over the rotation, and is also necessary as we might have multiple credentials from the source kind written to the same sink instance. For example, the same AWS Parameter Store sink might store AWS credentials from multiple AWS IAM users; the `key_to_name` mapping allows us to specify the names of the parameters rotator should update the value of for each source so that we don't overwrite the parameters for another source.

### Travis CI (`TravisCI`)
| Name | Description | Required |
|------|-------------|:-----:|
| repo\_slug | The target [Travis CI repository slug](https://developer.travis-ci.com/resource/env_var). Same as {repository.owner.name}/{repository.name}. | yes |

[`TRAVIS_API_AUTH_TOKEN`](https://github.com/shuheiktgw/go-travis#authentication-with-travis-api-token) should be set.

### AWS Systems Manager Parameter Store (`AWSParameterStore`)
| Name | Description | Required |
|------|-------------|:-----:|
| role\_arn | The ARN of the AWS IAM role that rotator should assume. | yes |
| region | The [AWS Regional endpoint[(https://docs.aws.amazon.com/general/latest/gr/rande.html) | yes |
| external\_id | If set, the [external ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html) is passed to the AWS STS AssumeRole API to assume the IAM Role specified by `role_arn` e.g. if deploying rotator on Kubernetes. | no |

[AWS credentials must be specified](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials) using a shared credentials file or `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.

### AWS Secrets Manager  (`AWSSecretsManager`)
| Name | Description | Required |
|------|-------------|:-----:|
| role\_arn | The ARN of the AWS IAM role that rotator should assume. | yes |
| region | The [AWS Regional endpoint[(https://docs.aws.amazon.com/general/latest/gr/rande.html) | yes |
| external\_id | If set, the [external ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html) is passed to the AWS STS AssumeRole API to assume the IAM Role specified by `role_arn` e.g. if deploying rotator on Kubernetes. | no |

[AWS credentials must be specified](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials) using a shared credentials file or `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.

## Contributing

Contributions and ideas are welcome! Please see [our contributing guide](CONTRIBUTING.md) and don't hesitate to open an issue or send a pull request to improve the functionality of this gem.

This project adheres to the Contributor Covenant [code of conduct](https://github.com/chanzuckerberg/.github/tree/master/CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to opensource@chanzuckerberg.com.

## License

[MIT](https://github.com/chanzuckerberg/sorbet-rails/blob/master/LICENSE)
