VERSION ?= $(shell git describe --tags --always --dirty)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X github.com/jonwraymond/metatools-mcp/cmd/metatools/cmd.Version=$(VERSION)
LDFLAGS += -X github.com/jonwraymond/metatools-mcp/cmd/metatools/cmd.GitCommit=$(GIT_COMMIT)
LDFLAGS += -X github.com/jonwraymond/metatools-mcp/cmd/metatools/cmd.BuildDate=$(BUILD_DATE)

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o bin/metatools ./cmd/metatools

.PHONY: install
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/metatools

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -rf bin/
