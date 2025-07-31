# AWSM Makefile

# Variables
BINARY_NAME=awsm
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)"
GO_FILES=$(shell find . -name "*.go" -not -path "./vendor/*")
GOBIN=$(shell go env GOPATH)/bin
COVERAGE_DIR=./coverage
DOCS_DIR=./docs

# Default target
.PHONY: all
all: clean lint test build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/awsm

# Install the application
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	@mv $(BINARY_NAME) $(GOBIN)/$(BINARY_NAME)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(COVERAGE_DIR)
	@rm -rf dist/

# Run tests
.PHONY: test
test: test-unit test-integration

# Run unit tests
.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	@mkdir -p $(COVERAGE_DIR)
	@go test -race -coverprofile=$(COVERAGE_DIR)/unit.out ./internal/... ./cmd/...

# Run integration tests
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	@mkdir -p $(COVERAGE_DIR)
	@go test -race -coverprofile=$(COVERAGE_DIR)/integration.out ./tests/integration/...

# Run tests with coverage
.PHONY: test-coverage
test-coverage: test
	@echo "Generating coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	@go tool cover -html=$(COVERAGE_DIR)/unit.out -o $(COVERAGE_DIR)/unit.html
	@go tool cover -html=$(COVERAGE_DIR)/integration.out -o $(COVERAGE_DIR)/integration.html
	@echo "Coverage reports generated in $(COVERAGE_DIR)/"

# Run linters
.PHONY: lint
lint: lint-go lint-fmt lint-vet

# Run gofmt
.PHONY: lint-fmt
lint-fmt:
	@echo "Running gofmt..."
	@gofmt -l -s -w $(GO_FILES)

# Run go vet
.PHONY: lint-vet
lint-vet:
	@echo "Running go vet..."
	@go vet ./...

# Run golangci-lint
.PHONY: lint-go
lint-go:
	@echo "Running golangci-lint..."
	@if [ ! -f $(GOBIN)/golangci-lint ]; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@$(GOBIN)/golangci-lint run ./...

# Generate documentation
.PHONY: docs
docs: docs-godoc docs-markdown

# Generate godoc
.PHONY: docs-godoc
docs-godoc:
	@echo "Generating godoc..."
	@if [ ! -f $(GOBIN)/godoc ]; then \
		echo "Installing godoc..."; \
		go install golang.org/x/tools/cmd/godoc@latest; \
	fi
	@echo "Run 'godoc -http=:6060' and visit http://localhost:6060/pkg/github.com/ao/awsm/ to view the documentation"

# Generate markdown documentation
.PHONY: docs-markdown
docs-markdown:
	@echo "Generating markdown documentation..."
	@mkdir -p $(DOCS_DIR)
	@if [ ! -f $(GOBIN)/gomarkdoc ]; then \
		echo "Installing gomarkdoc..."; \
		go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest; \
	fi
	@$(GOBIN)/gomarkdoc --output $(DOCS_DIR)/{{.Dir}}.md ./...

# Build for all platforms
.PHONY: dist
dist: clean
	@echo "Building for all platforms..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/awsm
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/awsm
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/awsm
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/awsm
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/awsm
	@echo "Binaries built in dist/"

# Create release archives
.PHONY: release
release: dist
	@echo "Creating release archives..."
	@cd dist && tar -czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	@cd dist && tar -czf $(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	@cd dist && tar -czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	@cd dist && tar -czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	@cd dist && zip $(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "Release archives created in dist/"

# Generate release notes
.PHONY: release-notes
release-notes:
	@echo "Generating release notes..."
	@echo "## $(VERSION) ($(shell date +%Y-%m-%d))" > .release-notes.md
	@echo "" >> .release-notes.md
	@if [ -n "$(shell git tag -l --sort=-v:refname | head -n 2 | tail -n 1)" ]; then \
		echo "### Changes since $$(git tag -l --sort=-v:refname | head -n 2 | tail -n 1)" >> .release-notes.md; \
		echo "" >> .release-notes.md; \
		git log --pretty=format:"* %s" $$(git tag -l --sort=-v:refname | head -n 2 | tail -n 1)..HEAD >> .release-notes.md; \
	else \
		echo "### Changes" >> .release-notes.md; \
		echo "" >> .release-notes.md; \
		git log --pretty=format:"* %s" >> .release-notes.md; \
	fi
	@echo "" >> .release-notes.md
	@echo "Release notes generated at .release-notes.md"

# Create a new release
.PHONY: release-create
release-create: release-notes
	@echo "Creating release $(VERSION)..."
	@if [ -z "$(shell git tag -l v$(VERSION))" ]; then \
		git tag -a v$(VERSION) -m "Release v$(VERSION)"; \
		echo "Tagged v$(VERSION)"; \
		echo "Push the tag with: git push origin v$(VERSION)"; \
	else \
		echo "Tag v$(VERSION) already exists"; \
	fi

# Bump version (usage: make bump-version VERSION=x.y.z)
.PHONY: bump-version
bump-version:
	@echo "Bumping version to $(VERSION)..."
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make bump-version VERSION=x.y.z"; \
		exit 1; \
	fi
	@sed -i.bak 's/Version = "[^"]*"/Version = "$(VERSION)"/' version.go
	@rm -f version.go.bak
	@echo "Version bumped to $(VERSION) in version.go"

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME)

# Run the application in TUI mode
.PHONY: run-tui
run-tui: build
	@echo "Running $(BINARY_NAME) in TUI mode..."
	@./$(BINARY_NAME) mode tui

# Run the application in TUI mode (direct command)
.PHONY: run-tui-direct
run-tui-direct: build
	@echo "Running $(BINARY_NAME) in TUI mode (direct command)..."
	@./$(BINARY_NAME) tui

# Show help
.PHONY: help
help:
	@echo "AWSM Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all            Build, lint, and test the application (default)"
	@echo "  build          Build the application"
	@echo "  install        Install the application"
	@echo "  clean          Clean build artifacts"
	@echo "  test           Run all tests"
	@echo "  test-unit      Run unit tests"
	@echo "  test-integration Run integration tests"
	@echo "  test-coverage  Run tests with coverage"
	@echo "  lint           Run all linters"
	@echo "  lint-fmt       Run gofmt"
	@echo "  lint-vet       Run go vet"
	@echo "  lint-go        Run golangci-lint"
	@echo "  docs           Generate documentation"
	@echo "  docs-godoc     Generate godoc"
	@echo "  docs-markdown  Generate markdown documentation"
	@echo "  dist           Build for all platforms"
	@echo "  release        Create release archives"
	@echo "  run            Run the application"
	@echo "  run-tui        Run the application in TUI mode"
	@echo "  run-tui-direct Run the application in TUI mode (direct command)"
	@echo "  help           Show this help"