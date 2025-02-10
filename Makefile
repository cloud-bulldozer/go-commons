.PHONY: lint test unittest help all

UNITTEST_PKG ?= "./..."

all: test

help:
	@echo "Make targets:"
	@echo
	@echo 'Usage:'
	@echo '    make lint                     Execute pre-commit linters'
	@echo '    make unittest                 Execute unittest'
	@echo '    make help                     Show this message'

test: lint unittest

unittest:
	ginkgo -r --randomize-all --randomize-suites --fail-on-pending --cover --trace --v --coverprofile=coverage.out ${UNITTEST_PKG}

lint:
	@echo "Executing pre-commit for all files"
	pre-commit run --all-files
	@echo "pre-commit executed."
