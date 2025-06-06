# Regular colors
color_green   := $(shell printf "\e[32m")  # Green color
color_reset   := $(shell printf "\e[0m")   # Reset to default color

# Go commands
GO := go
GOBUILD := $(GO) build
GOCLEAN := $(GO) clean

# Binany name
APP_NAME := semver

# Directories
BUILD_DIR := build
CMD_DIR := cmd/$(APP_NAME)

# ==================================================================================== #
# HELPERS
# ==================================================================================== #
.PHONY: help
help: ## Print this help message
	@echo ""
	@echo "Usage: make [action]"
	@echo ""
	@echo "Available Actions:"
	@echo ""
	@awk -v green="$(color_green)" -v reset="$(color_reset)" -F ':|##' \
		'/^[^\t].+?:.*?##/ {printf " %s* %-15s%s %s\n", green, $$1, reset, $$NF}' $(MAKEFILE_LIST) | sort
	@echo ""

# ==================================================================================== #
# PUBLIC TASKS
# ==================================================================================== #
.PHONY: all
all: clean build

.PHONY: clean
clean: ## Clean the build directory and Go cache
	@echo "$(color_bold_cyan)* Clean the build directory and Go cache$(color_reset)"
	@rm -rf $(BUILD_DIR)
	$(GOCLEAN) -cache

.PHONY: test
test: ## Run all tests and generate coverage report.
	@echo "$(color_bold_cyan)* Run all tests and generate coverage report.$(color_reset)"
	@$(GO) test $(go list ./... | grep -Ev 'internal/testutils') -coverprofile=coverage.txt
	@echo "$(color_bold_cyan)* Total Coverage$(color_reset)"
	@$(GO) tool cover -func=coverage.txt | grep total | awk '{print $$3}'

.PHONY: test/force
test/force: ## Clean go tests cache.
	@echo "$(color_bold_cyan)* Clean go tests cache and run all tests.$(color_reset)"
	@$(GO) clean -testcache
	@$(MAKE) test

.PHONY: modernize
modernize: ## Run go-modernize
	@echo "$(color_bold_cyan)* Running go-modernize$(color_reset)"
	@modernize -test ./...

.PHONY: lint
lint: ## Run golangci-lint
	@echo "$(color_bold_cyan)* Running golangci-lint$(color_reset)"
	@golangci-lint run ./...

.PHONY: build
build: ## Build the binary with development metadata
	@echo "$(color_bold_cyan)* Building the binary...$(color_reset)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) ./$(CMD_DIR)

.PHONY: install
install: ## Install the binary using Go install
	@$(MAKE) modernize
	@$(MAKE) lint
	@$(MAKE) test/force
	@echo "$(color_bold_cyan)* Install the binary using Go install$(color_reset)"
	@cd $(CMD_DIR) && $(GO) install .

# catch-all rule: do nothing for undefined targets instead of throwing an error
%:
	@:
