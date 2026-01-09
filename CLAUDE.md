# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Compile a Rugby file to Go
go run . <file.rby>

# Run the generated Go code
go run <file.go>

# Build the compiler binary
go build -o rugby .

# Run tests (excludes examples/ which has multiple main packages)
go test ./lexer/... ./parser/... ./codegen/...
```

## Architecture

Rugby is a compiler that transforms Ruby-like syntax into idiomatic Go code. The compilation pipeline:

```
Source (.rby) → Lexer → Parser → AST → CodeGen → Go (.go)
```

**Packages:**

- `token/` - Token types (IMPORT, DEF, END, IDENT, STRING, etc.)
- `lexer/` - Tokenizer that handles Ruby-style syntax (newline-significant, `#` comments, `do/end` blocks)
- `ast/` - AST node definitions (Program, FuncDecl, CallExpr, etc.)
- `parser/` - Recursive descent parser
- `codegen/` - Go code emitter that maps Rugby constructs to Go (e.g., `puts` → `fmt.Println`)
- `main.go` - CLI entry point

## Language Spec

See `spec.md` for the full Rugby language specification. Key points:

- Ruby-like surface syntax, Go-like semantics
- Compiles to readable Go code
- No metaprogramming, no monkey patching, no eval
- First-class Go interop (snake_case → CamelCase mapping)
- `T?` optionals lower to Go's `(T, bool)` pattern

## Rules
* Commit messages should be concise and never include stats (number of files, changes, etc) or attribution to claude or claude code.
