.PHONY: build install clean test

build:
	go build -o flux-relay .

install: build
	go install .

clean:
	rm -f flux-relay flux-relay.exe

test:
	go test ./...

deps:
	go mod download
	go mod tidy
