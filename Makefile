BINARY := eng-graph
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X github.com/eng-graph/eng-graph/cmd.version=$(VERSION)"

.PHONY: build install test lint clean

build:
	go build $(LDFLAGS) -o bin/$(BINARY) .

install:
	go install $(LDFLAGS) .

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
