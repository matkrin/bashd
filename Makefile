BINNAME := bashd
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
LDFLAGS := "-X 'main.VERSION=$(VERSION)'"


build:
	go build -ldflags=$(LDFLAGS) -o bin/$(BINNAME) cmd/$(BINNAME)/main.go

test:
	@go test ./...
