# Go commands
go := "go"
gobuild := go + " build"
goclean := go + " clean"

# Binary name
app_name := "verso"

# Directories
build_dir := "build"
cmd_dir := "cmd/" + app_name

# Default recipe: show help
default: help

# Print this help message
help:
    @echo ""
    @echo "Usage: just [recipe]"
    @echo ""
    @echo "Available Recipes:"
    @echo ""
    @just --list
    @echo ""

# Clean and build
all: clean build

# Clean the build directory and Go cache
clean:
    @echo "* Clean the build directory and Go cache"
    rm -rf {{ build_dir }}
    {{ goclean }} -cache

# Run go-modernize with auto-fix
modernize:
    @echo "* Running go-modernize"
    modernize --fix ./...

# Run golangci-lint
lint:
    @echo "* Running golangci-lint"
    golangci-lint run ./...

# Run goreportcard-cli
reportcard:
    @echo "* Running goreportcard-cli..."
    goreportcard-cli -v

# Run all tests and generate coverage report
test:
    @echo "* Run all tests and generate coverage report."
    {{ go }} test $({{ go }} list ./... | grep -Ev 'internal/testutils') -coverprofile=coverage.txt
    @echo "* Total Coverage"
    {{ go }} tool cover -func=coverage.txt | grep total | awk '{print $3}'

# Clean go tests cache and run all tests
test-force:
    @echo "* Clean go tests cache and run all tests."
    {{ go }} clean -testcache
    just test

# Run modernize, lint, and reportcard
check: modernize lint reportcard

# Build the binary with development metadata
build:
    @echo "* Building the binary..."
    mkdir -p {{ build_dir }}
    {{ gobuild }} -o {{ build_dir }}/{{ app_name }} ./{{ cmd_dir }}

# Install the binary using Go install
install: check test-force
    @echo "* Install the binary using Go install"
    cd {{ cmd_dir }} && {{ go }} install .
