BINNAME := bashd
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
LDFLAGS := "-X 'main.VERSION=$(VERSION)'"

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
EXT :=

ifeq ($(GOOS),windows)
    EXT := .exe
endif

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags=$(LDFLAGS) -o bin/$(BINNAME)$(EXT) cmd/$(BINNAME)/main.go

watch:
	./scripts/watch.sh

test:
	@go test ./...
