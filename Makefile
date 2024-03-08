LINTER_VERSION=1.56.2
GCI_VERSION := latest

.PHONY: all
all: lint test

.PHONY: lint
lint: install-lint
	@echo 'run golangci lint'
	golangci-lint run --out-format=tab

.PHONE: install-lint
install-lint:
	@if ! golangci-lint --version | grep -q $(LINTER_VERSION); \
		then curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v$(LINTER_VERSION); fi;

.PHONY: test
test:
	@echo 'running tests'
	go test -v -cover -race ./...

.PHONY: format
format: install-gci
	@echo 'format code and imports'
	@go fmt ./...
	@gci write . --skip-generated -s standard -s default -s 'Prefix(github.com/HardDie)' -s 'Prefix(github.com/HardDie/fsentry)'

.PHONE: install-gci
install-gci:
	@if ! gci --version | grep -q gci; \
		then go install github.com/daixiang0/gci@$(GCI_VERSION); fi;
