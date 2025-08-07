# Makefile for building the MCPFier project

APP_NAME = mcpfier

.PHONY: all build test clean

all: build

# Build the Go application
build:
	go build -o $(APP_NAME)

# Run tests
 test:
	go test ./...

# Clean the build artifacts
clean:
	rm -f $(APP_NAME)
