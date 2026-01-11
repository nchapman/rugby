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
```

Direct commands (without make):

```bash
# Run a Rugby file directly
go run . examples/hello.rg

# Run a single test file
go test ./parser -run TestFunctionCall

# Run tests with coverage
go test -cover ./...
```

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
- Handles `?` and `!` suffixes on identifiers (e.g., `empty?`, `save!`)
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
- `mapType()`: `Int` → `int`, `String?` → `runtime.OptionalString`

**Kernel functions** (auto-mapped): `puts`, `print`, `p`, `gets`, `exit`, `sleep`, `rand`

**Block methods** (→ runtime calls): `each`, `map`, `select`, `reject`, `find`, `reduce`, `any?`, `all?`, `none?`, `times`, `upto`, `downto`

### runtime/*.go
Go package providing Ruby-like methods, imported as `rugby/runtime`.
- **array.go**: `Each`, `Map`, `Select`, `Reject`, `Find`, `Reduce`, `Any`, `All`, `None`, `Contains`, `First`, `Last`
- **map.go**: `MapEach`, `Keys`, `Values`, `Fetch`, `Merge`
- **string.go**: `StringToInt`, `StringToFloat`, `Chars`, `StringReverse`
- **int.go**: `Times`, `Upto`, `Downto`, `Abs`, `Clamp`
- **range.go**: `Range` struct, `RangeEach`, `RangeContains`, `RangeToArray`, `RangeSize`
- **optional.go**: `OptionalInt`, `OptionalString`, etc. with `SomeT()`/`NoneT()` constructors
- **equal.go**: `Equal()` for deep equality (checks `Equaler` interface first)
- **io.go**: `Puts`, `Print`, `P`, `Gets`, `Exit`, `Sleep`

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
