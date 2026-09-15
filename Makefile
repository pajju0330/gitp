.PHONY: build install test clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X github.com/prajwal/gitp/internal/app.Version=$(VERSION)"
PREFIX  ?= $(or $(GOBIN),$(shell go env GOPATH)/bin)

build:
	go build $(LDFLAGS) -o bin/gitp ./cmd/gitp

install: build
	mkdir -p "$(PREFIX)"
	install -m 755 bin/gitp "$(PREFIX)/gitp"

test:
	go test ./...

clean:
	rm -rf bin
