.PHONY: build clean test install

# Build the binary
build:
	go build -o version-policy main.go

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o version-policy-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o version-policy-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o version-policy-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -o version-policy-windows-amd64.exe main.go

# Run tests
test:
	go test -v ./...

# Install binary locally
install: build
	cp version-policy $(HOME)/.local/bin/

# Clean build artifacts
clean:
	rm -f version-policy version-policy-* *.exe

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Default target
all: clean build
