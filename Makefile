GO_BIN := $(GOPATH)/bin
GOIMPORTS := go run golang.org/x/tools/cmd/goimports@latest
GOFUMPT := go run mvdan.cc/gofumpt@v0.3.1
GOLINT := $(GO_BIN)/golangci-lint

all: install

install: fmt
	go install -v

test:
	go test ./... -v

coverage:
	go test ./... -v -race -coverprofile=coverage.txt -covermode=atomic

lint: $(GOLINT)
	golangci-lint run

fmt:
	$(GOIMPORTS) -w -l .
	$(GOFUMPT) -w -l .

$(GOIMPORTS):
	go get -u golang.org/x/tools/cmd/goimports

$(GOLINT):
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: install test fmt lint
