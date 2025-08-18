.PHONY: lint test unittest help all build-cli

UNITTEST_PKG ?= "./..."

all: test

help:
	@echo "Make targets:"
	@echo
	@echo 'Usage:'
	@echo '    make lint                     Execute pre-commit linters'
	@echo '    make unittest                 Execute unittest'
	@echo '    make build-cli                Build the ocp-metadata CLI tool'
	@echo '    make help                     Show this message'

test: lint unittest

unittest:
	ginkgo -r --randomize-all --randomize-suites --fail-on-pending --cover --trace --v --coverprofile=coverage.out ${UNITTEST_PKG}

lint:
	@echo "Executing pre-commit for all files"
	pre-commit run --all-files
	@echo "pre-commit executed."

build-cli:
	@echo "Building ocp-metadata CLI tool"
	go build -o bin/ocp-metadata ./cmd/ocp-metadata
	@echo "CLI tool built at bin/ocp-metadata"
