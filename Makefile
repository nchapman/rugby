.PHONY: build test test-spec test-spec-bless lint clean install run repl help

# Build the rugby compiler
build:
	@echo "Building rugby compiler..."
	@go build -o rugby .

# Run all tests
test:
	@echo "Running tests..."
	@go test ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./...

# Run spec tests (language specification tests)
test-spec: build
	@echo "Running spec tests..."
	@cd tests && go test -v -run TestSpecs

# Run spec tests and update golden files
test-spec-bless: build
	@echo "Running spec tests with blessing..."
	@cd tests && go test -v -run TestSpecs -args -bless

# Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	@golangci-lint fmt ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f rugby
	@rm -rf .rugby/

# Install rugby to $GOPATH/bin
install: build
	@echo "Installing rugby..."
	@go install .

# Run the REPL
repl: build
	@./rugby repl

# Run a specific Rugby file (use: make run FILE=examples/hello.rg)
run: build
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make run FILE=path/to/file.rg"; \
		exit 1; \
	fi
	@./rugby run $(FILE)

# Run all checks (test + lint)
check: test lint
	@echo "All checks passed!"

# Build a Rugby file to binary (use: make build-rugby FILE=main.rg OUTPUT=myapp)
build-rugby: build
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make build-rugby FILE=main.rg OUTPUT=myapp"; \
		exit 1; \
	fi
	@./rugby build $(FILE) -o $(OUTPUT)

# Display help
help:
	@echo "Rugby Makefile targets:"
	@echo "  build           - Build the rugby compiler"
	@echo "  test            - Run all tests"
	@echo "  test-coverage   - Run tests with coverage"
	@echo "  test-spec       - Run language specification tests"
	@echo "  test-spec-bless - Run spec tests and update golden files"
	@echo "  lint            - Run static analysis"
	@echo "  fmt             - Format code"
	@echo "  clean           - Clean build artifacts"
	@echo "  install         - Install rugby to GOPATH/bin"
	@echo "  repl            - Start the Rugby REPL"
	@echo "  run             - Run a Rugby file (use: make run FILE=examples/hello.rg)"
	@echo "  build-rugby     - Build a Rugby file to binary (use: make build-rugby FILE=main.rg OUTPUT=myapp)"
	@echo "  check           - Run tests and linters"
	@echo "  help            - Show this help message"
