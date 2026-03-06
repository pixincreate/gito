.PHONY: build install test clean

BINARY_NAME=gito
VERSION=0.1.0
BUILD_DIR=dist

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags="-s -w" .

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)

fmt:
	go fmt ./...

lint:
	golangci-lint run

deps:
	go mod download
	go mod tidy
