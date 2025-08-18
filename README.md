# go-commons

[![codecov](https://codecov.io/github/cloud-bulldozer/go-commons/branch/main/graph/badge.svg?token=CFPW1UV7FO)](https://codecov.io/github/cloud-bulldozer/go-commons)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloud-bulldozer/go-commons)](https://goreportcard.com/report/github.com/cloud-bulldozer/go-commons)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Running test locally
Use `make` to run lint and unittest locally. See `make help` for details

### `lint`
The `lint` tests are executed using [pre-commit](https://pre-commit.com/).
Make sure to install it before running.

### `unittest`
The `unittest` tests are executed using [ginkgo](https://onsi.github.io/ginkgo/).
Make sure to install it locally before running.

### `build-cli`
Will build a simple CLI tool to retrieve the cluster metadata.