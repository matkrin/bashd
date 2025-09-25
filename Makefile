BINNAME := bashd

build:
	@go build -o bin/$(BINNAME) cmd/$(BINNAME)/main.go

test:
	@go test ./...
