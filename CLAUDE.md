# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

Common commands are available via `make`:

```bash
# See all available targets
make help

# Build the rugby compiler
make build

# Run tests and linters
make check

# Run all tests
make test

# Run spec tests (language feature tests)
make test-spec

# Run linters
make lint

# Format code
make fmt

# Start the interactive REPL
make repl

# Run a Rugby file
make run FILE=examples/hello.rg

# Build a standalone binary from a Rugby file
make build-rugby FILE=main.rg OUTPUT=myapp

# Clean build artifacts
make clean

# Clear all .rugby cache directories
make clean-cache
```

### Running Single Tests

**IMPORTANT:** Always use make targets for spec tests. They rebuild the compiler and clear caches to avoid stale binaries/generated code.

```bash
# Spec tests - always use make (rebuilds compiler, clears caches)
make test-spec                            # all spec tests
make test-spec NAME=interfaces/any_type   # single test
make test-spec NAME=blocks                # category

# Unit tests
make test                                 # all unit tests
make test PKG=codegen                     # package
make test PKG=codegen TEST=TestClass      # specific test
```

**DO NOT run `go test` directly for spec tests** - it won't rebuild the compiler or clear caches, leading to confusing failures from stale code.

## Architecture

Rugby is a compiler that transforms Ruby-like syntax into idiomatic Go code. The compilation pipeline:

```
Source (.rg) → Lexer → Parser → AST → CodeGen → Go (.go) → go build → Binary
```

**Core packages:**

- `token/` - Token types (IMPORT, DEF, END, IDENT, STRING, etc.)
- `lexer/` - Tokenizer that handles Ruby-style syntax (newline-significant, `#` comments, `do/end` blocks)
- `ast/` - AST node definitions (Program, FuncDecl, CallExpr, etc.)
- `parser/` - Recursive descent parser
- `codegen/` - Go code emitter that maps Rugby constructs to Go (e.g., `puts` → `runtime.Puts`)
- `runtime/` - Go runtime library providing Ruby-like methods (array iteration, string methods, etc.)

**CLI and tooling:**

- `cmd/` - Cobra CLI commands (run, build, clean, repl)
- `internal/builder/` - Orchestrates compilation: .rg → .go → binary, manages `.rugby/` directory
- `internal/repl/` - Interactive REPL using bubbletea

## Package Internals

### token/token.go
Defines `TokenType` constants and `Token` struct with `Type`, `Literal`, `Line`, `Column`. Keywords are mapped via `LookupIdent()`. Key tokens: `DEF`, `END`, `CLASS`, `IF`, `ELSIF`, `FOR`, `IN`, `DO`, `PUB`, `SELF`, `ARROW` (->), `DOTDOT` (..), `TRIPLEDOT` (...), `ORASSIGN` (||=).

### lexer/lexer.go
Single-pass tokenizer. Entry point: `New(input string)` → `*Lexer`, then call `NextToken()` repeatedly.
- Newlines are significant tokens (Ruby-style)
- `#` starts comments (skipped)
- Handles `?` suffix on identifiers for predicate methods (e.g., `empty?`, `valid?`)
- `SaveState()`/`RestoreState()` for parser lookahead

### ast/ast.go
All nodes implement `Node` interface. Two sub-interfaces: `Statement` and `Expression`.
- **Program**: root node with `Imports []*ImportDecl` and `Declarations []Statement`
- **Key statements**: `FuncDecl`, `ClassDecl`, `InterfaceDecl`, `IfStmt`, `WhileStmt`, `ForStmt`, `AssignStmt`, `ReturnStmt`
- **Key expressions**: `CallExpr` (with optional `Block *BlockExpr`), `SelectorExpr`, `BinaryExpr`, `RangeLit`, `InstanceVar`
- Classes have `Fields []*FieldDecl` (inferred from `initialize`) and `Methods []*MethodDecl`

### parser/parser.go
Recursive descent with Pratt parsing for expressions. Entry: `New(lexer) → ParseProgram() *ast.Program`.

**Key patterns:**
- `parseStatement()` dispatches on token type to `parseIfStmt()`, `parseForStmt()`, etc.
- `parseExpression(precedence)` is the Pratt parser core - handles prefix/infix with precedence climbing
- `parseBlockCall()` handles `do |x| ... end` and `{ |x| ... }` blocks attached to method calls
- `extractFields()` infers struct fields from `@var = param` assignments in `initialize`

**Precedence levels** (low to high): RANGE → OR → AND → EQUALS → LESSGREATER → SUM → PRODUCT → PREFIX → CALL → MEMBER

### codegen/codegen.go
Single-pass emitter. Entry: `New() → Generate(program) (string, error)`.

**Key patterns:**
- `genStatement()` dispatches to `genFuncDecl()`, `genClassDecl()`, `genIfStmt()`, etc.
- `genExpr()` dispatches to `genCallExpr()`, `genBinaryExpr()`, `genSelectorExpr()`, etc.
- `genBlockCall()` handles Ruby blocks → Go closures with `runtime.*` calls
- Tracks `vars map[string]string` for declared variables and types
- Tracks `imports map[string]bool` to detect Go interop (PascalCase vs camelCase)
- `needsRuntime bool` triggers automatic `rugby/runtime` import

**Name transformations:**
- `snakeToCamelWithAcronyms()`: `user_id` → `userID` (private methods)
- `snakeToPascalWithAcronyms()`: `user_id` → `UserID` (exported, Go interop)
- `mapType()`: `Int` → `int`, `String?` → `*string` (pointer-based optionals)

**Kernel functions** (auto-mapped): `puts`, `print`, `p`, `gets`, `exit`, `sleep`, `rand`

**Block methods** (→ runtime calls): `each`, `map`, `select`, `reject`, `find`, `reduce`, `any?`, `all?`, `none?`, `times`, `upto`, `downto`

### runtime/*.go
Go package providing Ruby-like methods, imported as `rugby/runtime`.
- **array.go**: `Each`, `Map`, `Select`, `Reject`, `Find`, `Reduce`, `Any`, `All`, `None`, `Contains`, `First`, `Last`
- **map.go**: `MapEach`, `Keys`, `Values`, `Fetch`, `Merge`
- **string.go**: `StringToInt`, `StringToFloat`, `Chars`, `StringReverse`
- **int.go**: `Times`, `Upto`, `Downto`, `Abs`, `Clamp`
- **float.go**: `FloatAbs`, `FloatRound`, `FloatFloor`, `FloatCeil`
- **range.go**: `Range` struct, `RangeEach`, `RangeContains`, `RangeToArray`, `RangeSize`
- **optional.go**: Generic `Some[T]()`, `Coalesce[T]()` for pointer-based optionals (`*T`); `Option[T]` struct for field storage
- **set.go**: `Set[T]` generic set type with `Add`, `Contains`, `Remove`, `ToSlice`
- **task.go**: `Task[T]` for spawn/await concurrency pattern
- **equal.go**: `Equal()` for deep equality (checks `Equaler` interface first)
- **io.go**: `Puts`, `Print`, `P`, `Gets`, `Exit`, `Sleep`

## Spec Testing

The `tests/spec/` directory contains spec-driven tests that verify language features end-to-end. These tests catch integration bugs that unit tests miss.

```bash
# Run spec tests
make test-spec

# Run spec tests and update golden files
make test-spec-bless
```

### Test File Format

Spec tests are self-contained `.rg` files with directives:

```ruby
#@ run-pass           # Must compile and run successfully
#@ check-output       # Verify stdout matches expected

puts "Hello"

#@ expect:
# Hello
```

### Directives

| Directive | Meaning |
|-----------|---------|
| `#@ run-pass` | Must compile and run without error (default) |
| `#@ run-fail` | Must compile but fail at runtime |
| `#@ compile-fail` | Must fail to compile |
| `#@ check-output` | Verify stdout matches `#@ expect:` block |
| `#@ skip: reason` | Skip test with reason |
| `#~ ERROR: pattern` | Inline: expected error matching regex |

### Directory Structure

```
tests/spec/
├── literals/          # Integer, string, array, range literals
├── types/             # Type annotations, generics, type inference
├── functions/         # Function definitions, multiple returns
├── classes/           # Class definitions, inheritance
├── structs/           # Struct definitions, value types
├── interfaces/        # Interface definitions, structural typing
├── modules/           # Module definitions, namespacing
├── enums/             # Enum types
├── control_flow/      # If/elsif/else, case, loops
├── blocks/            # Block iteration (each, map, select)
├── optionals/         # Optional types, if-let, nil coalescing
├── error_handling/    # Error propagation (!), rescue
├── generics/          # Generic types and functions
├── go_interop/        # Go package imports
├── concurrency/       # Channels, spawn/await
├── runtime/           # Runtime library features
├── stdlib/            # Standard library tests
├── visibility/        # pub/private visibility
├── integration/       # Multi-feature integration tests
└── errors/            # compile-fail tests for known bugs
```

### Writing New Spec Tests

1. Create a `.rg` file in the appropriate category
2. Add directives at the top (`#@ run-pass`, etc.)
3. For expected output, use `#@ expect:` block or create a `.stdout` golden file
4. Run `make test-spec` to verify

### Bug Documentation

Known bugs are documented as `compile-fail` tests in `tests/spec/errors/`. When a bug is fixed, the test is moved to the appropriate feature directory and converted to `run-pass`.

## Language Spec

See `spec.md` for the full Rugby language specification. Key points:

- Ruby-like surface syntax, Go-like semantics
- Compiles to readable Go code
- No metaprogramming, no monkey patching, no eval
- First-class Go interop (snake_case → CamelCase mapping)
- `T?` optionals lower to Go's `(T, bool)` pattern

## Rules
* Commit messages should be concise and never include stats (number of files, changes, etc) or attribution to claude or claude code.
* Always add tests for new features (lexer, parser, and codegen tests as appropriate).
* Run `make check` before committing to ensure tests and linters pass.
* Always run the code-reviewer agent before committing to catch issues early.
