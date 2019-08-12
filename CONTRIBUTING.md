# Contributing to Rotator

Thank you for taking the time to contribute to this project!

This project adheres to the Contributor Covenant
[code of conduct](https://github.com/chanzuckerberg/.github/tree/master/CODE_OF_CONDUCT.md).
By participating, you are expected to uphold this code. Please report unacceptable behavior
to opensource@chanzuckerberg.com.

This project is licensed under the [MIT license](LICENSE.md).

## Reporting Bugs and Adding Functionality

We're excited you'd like to contribute to rotator!

When reporting a bug, please include:
 * Steps to reproduce
 * The version of rotator that you are using
 * A test case, if you are able

**If you believe you have found a security issue, please contact us at security@chanzuckerberg.com**
rather than filing an issue here.

When proposing new functionality, please include tests that cover the new behavior.

## Local Development

0. Install go
1. Clone `rotator` locally:

```sh
❯ git clone https://github.com/chanzuckerberg/rotator.git
```
2. Set up development dependencies 
```sh
❯ make setup
```

## Tests

### Running Tests

You can run tests using `make test`.

To include integration tests, use `make test-all` instead. Integration tests rely on any external resources (e.g. AWS services) while unit tests rely on mocking.

To get accurate test coverage using [goverage](https://github.com/haya14busa/goverage), use `make test-coverage` or `make test-coverage-all`.
