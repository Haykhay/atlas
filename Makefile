BINARY  := atlas
VERSION ?= dev
LDFLAGS := -s -w -X github.com/Haykhay/atlas/internal/cli.Version=$(VERSION)

.PHONY: build test lint cover

build:
	go build -trimpath -ldflags '$(LDFLAGS)' -o $(BINARY) ./cmd/atlas

test:
	go test ./... -race

cover:
	go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -1

lint:
	golangci-lint run
