.PHONY: build run test clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -X github.com/white-blue-protocol/wblue/internal/version.Version=$(VERSION) \
           -X github.com/white-blue-protocol/wblue/internal/version.GitCommit=$(COMMIT) \
           -X github.com/white-blue-protocol/wblue/internal/version.BuildTime=$(BUILD_TIME)

build:
	go build -ldflags "$(LDFLAGS)" -o wblue .

run: build
	./wblue start

test:
	go test ./...

clean:
	rm -f wblue
	rm -rf ~/.wblue/data
