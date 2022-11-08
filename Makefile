GO_BIN := $(GOPATH)/bin
GOIMPORTS := go run golang.org/x/tools/cmd/goimports@latest
GOFUMPT := go run mvdan.cc/gofumpt@v0.3.1
GOLANGCI := go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1

.PHONY: all
all: build

.PHONY: build
build:
	go build

.PHONY: install
install:
	go install -v

.PHONY: test
test:
	go test ./...

.PHONY: coverage
coverage:
	go test ./... -v -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: lint
lint:
	$(GOLANGCI) run

.PHONY: fmt
fmt:
	$(GOIMPORTS) -w -l .
	$(GOFUMPT) -w -l .
