.PHONY: build test test-spec test-spec-bless lint clean clean-cache install run repl benchmark help

# Build the rugby compiler
build:
	@echo "Building rugby compiler..."
	@go build -o rugby .

# Run unit tests
# Usage: make test                              # run all
#        make test PKG=codegen                  # run all in package
#        make test PKG=codegen TEST=TestClass   # run specific test
test:
	@if [ -n "$(PKG)" ] && [ -n "$(TEST)" ]; then \
		go test -v ./$(PKG) -run $(TEST); \
	elif [ -n "$(PKG)" ]; then \
		go test -v ./$(PKG); \
	else \
		go test ./...; \
	fi

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./...

# Run spec tests (rebuilds compiler, clears caches)
# Usage: make test-spec                         # run all
#        make test-spec NAME=interfaces         # run category
#        make test-spec NAME=interfaces/any_type  # run specific test
test-spec: build clean-cache
	@if [ -n "$(NAME)" ]; then \
		cd tests && go test -v -run "TestSpecs/$(NAME)"; \
	else \
		cd tests && go test -v -run TestSpecs; \
	fi

# Run spec tests and update golden files
test-spec-bless: build clean-cache
	@echo "Running spec tests with blessing..."
	@cd tests && go test -v -run TestSpecs -args -bless

# Clean all .rugby cache directories (generated Go files)
clean-cache:
	@echo "Clearing .rugby caches..."
	@find . -name ".rugby" -type d -exec rm -rf {} + 2>/dev/null || true

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

# Run performance benchmarks (rebuilds compiler and all binaries for fair comparison)
benchmark: build clean-cache
	@echo "Cleaning all benchmark binaries..."
	@find benchmarks -name "*_rg" -type f -delete 2>/dev/null || true
	@find benchmarks -name "*_go" -type f -delete 2>/dev/null || true
	@find benchmarks -name "*_rs" -type f -delete 2>/dev/null || true
	@find benchmarks -name "*_cr" -type f -delete 2>/dev/null || true
	@cd benchmarks && ./bench.sh

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
	@echo ""
	@echo "Testing:"
	@echo "  make test                                 - Run all unit tests"
	@echo "  make test PKG=codegen                     - Run tests in a package"
	@echo "  make test PKG=codegen TEST=TestClass      - Run a specific test"
	@echo "  make test-spec                            - Run all spec tests"
	@echo "  make test-spec NAME=interfaces            - Run spec tests in category"
	@echo "  make test-spec NAME=interfaces/any_type   - Run a specific spec test"
	@echo "  make test-spec-bless                      - Update golden files"
	@echo "  make check                                - Run tests + linters"
	@echo ""
	@echo "Building:"
	@echo "  make build                                - Build the compiler"
	@echo "  make install                              - Install to GOPATH/bin"
	@echo "  make clean                                - Remove build artifacts"
	@echo "  make clean-cache                          - Clear .rugby caches"
	@echo ""
	@echo "Running:"
	@echo "  make run FILE=examples/hello.rg           - Run a Rugby file"
	@echo "  make build-rugby FILE=main.rg OUTPUT=app  - Build to binary"
	@echo "  make repl                                 - Start the REPL"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint                                 - Run linters"
	@echo "  make fmt                                  - Format code"
	@echo ""
	@echo "Benchmarks:"
	@echo "  make benchmark                            - Run performance benchmarks"
