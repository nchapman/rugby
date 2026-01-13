# Rugby Compiler Roadmap & TODOs

This document outlines the architectural roadmap for the Rugby compiler, prioritizing correctness, robustness, and performance.

## Phase 1: Semantic Analysis & Type Safety (High Priority)

Currently, codegen does too much: type inference, error checking, and generation. We need to separate these concerns to fail earlier and clearer.

- [x] **Symbol Table Implementation** ✅
    - Created `semantic` package with `types.go`, `symbol.go`, `scope.go`, `errors.go`, `analyzer.go`.
    - Implemented `Scope` structure to track variable declarations, function signatures, and class definitions.
    - Handle nesting (blocks, methods, classes) with proper scope chains.

- [x] **Type Checker / Inference Pass** ✅
    - Created two-pass analyzer that traverses the AST *before* codegen.
    - Annotate AST nodes with resolved types via `nodeTypes` map.
    - Implemented inference for:
        - [x] Variable assignments (`x = 5` → `x` is `Int`).
        - [x] Function return types (from declarations).
        - [x] Generic containers (infer `Array[Int]`, `Map[String, Int]`).

- [x] **Compile-Time Error Reporting** ✅
    - Integrated semantic analysis into build pipeline (`builder.go`).
    - Moved logic from `codegen` to `semantic` for:
        - [x] Undefined variables/methods.
        - [x] Argument count mismatches (with variadic builtin support).
        - [x] Type mismatches (e.g., `Int + String`) - binary expression and comparison type checking.
    - Provides rich error messages with "Did you mean?" suggestions.

- [x] **Extended Type Checking** ✅
    - [x] Unary expression type checking (`not 5` errors, `not` requires Bool, `-` requires numeric).
    - [x] Composite type equality (arrays/maps of same element type are comparable, different types error).
    - [x] Class/interface instance equality validation (same class comparable, different classes error).
    - [x] Class constructor type inference (`Dog.new()` returns `Dog` type).

- [x] **Comprehensive Type Validation** ✅
    - [x] Return type validation (returned value must match declared return type).
    - [x] Function argument type validation (arguments must match parameter types).
    - [x] Conditional expression validation (`if`/`while`/`until`/ternary conditions must be Bool).
    - [x] Array/string index type validation (index must be Int or Int64).
    - [x] Compound assignment type checking (`+=` for numerics/strings, `-=`/`*=`/`/=`/`%=` for numerics).
    - [x] Array/map element type homogeneity (all elements must have compatible types).
    - [x] Case/when value type matching (when values must be comparable to subject type).
    - [x] Channel send type validation (sent value must match channel element type).
    - [x] Nil coalesce type compatibility (`??` right side must be assignable to unwrapped left type).
    - [x] Test statement analysis (`describe`, `it`, `test`, `table`, `before`, `after` bodies analyzed).

- [x] **Method & Interface Type Checking** ✅
    - [x] Method call type checking (added `lookupMethod` with built-in method signatures for String, Int, Float, Array, Map, Range, Optional).
    - [x] Interface implementation validation (verify class implements all interface methods with matching signatures).
    - [x] Bang operator validation (`!` verifies expression is error tuple and function returns error).

- [x] **Remaining Type Checking** ✅
    - [x] Block parameter type inference (infers types from receiver: `Array[T].each` → block param is `T`).
    - [x] Safe navigation return type inference (`&.` now returns `Optional[T]` based on field/method type).
    - Note: Method visibility (`pub`) controls package exports (Go-style), not Ruby-style private methods.

## Phase 1.5: Testing & Development UX (Critical for Refactoring)

Testing generated string output is fragile, and single-error stopping slows down development. We need behavioral tests and better error reporting to refactor `codegen` with confidence.

- [x] **Integration Test Harness** ✅
    - [x] Harness exists in `examples/examples_test.go` - runs `rugby build`, executes binaries, asserts `stdout`/`stderr`/`exit_code`.
    - [x] Covers hello, arith, fizzbuzz, case, ranges, symbols, errors, json, http examples.

- [x] **Parser Error Recovery** ✅
    - [x] Parser already reports multiple errors per file (collects errors in slice, continues parsing).
    - [x] Added `synchronize()` and `syncToEnd()` utilities for future recovery improvements.
    - [x] Test added verifying multiple errors reported in one parse pass.

## Phase 1.75: Error Experience (Rust/Elm-Quality Errors)

Goal: Compiler errors should be helpful teachers, not cryptic complaints. Inspired by Rust and Elm's famously friendly error messages.

### Error Rendering System ✅

- [x] **Source Context Display**
    - [x] Store source code reference in compiler context (pass through lexer → parser → semantic).
    - [x] Display 1-3 lines of source code with each error.
    - [x] Show line numbers in left gutter (dimmed).
    - [x] Handle edge cases: first/last lines, very long lines (truncate with `...`).

- [x] **Precise Error Spans**
    - [x] Audit AST nodes to ensure all have start position (Line, Column).
    - [ ] Add end position (EndLine, EndColumn) to AST nodes that span multiple tokens.
    - [ ] Update lexer to track token end positions.
    - [x] Render underlines (`^~~~`) or carets (`^`) pointing to exact error location.
    - [ ] Support multi-line spans for things like unclosed blocks.

- [x] **Terminal Formatting**
    - [x] Create `errors/formatter.go` with `FormatError(err, source) string`.
    - [x] Add color support (detect TTY, respect `NO_COLOR` env var):
        - Red: error header and primary underline.
        - Cyan: help suggestions.
        - Yellow: warnings.
        - Blue: notes and secondary locations.
        - Dim: line numbers, context lines.
    - [x] Bold for error message text.
    - [x] Support `--color=always|never|auto` flag.

### Rich Error Information

- [ ] **Multi-Part Error Messages**
    - [ ] Primary message: What went wrong.
    - [ ] Secondary locations: "note: `x` was declared here" with source snippet.
    - [ ] Help text: Actionable suggestion for how to fix.
    - [ ] Update error structs to support `Notes []Note` and `Help string` fields.

- [ ] **Contextual Help Suggestions**
    - [ ] Type mismatches: "help: consider converting with `.to_i` or `.to_s`".
    - [ ] Missing methods: "help: did you forget to define `initialize`?".
    - [ ] Wrong argument count: "help: this function expects 2 arguments: `(name: String, age: Int)`".
    - [ ] Undefined with typo: "help: a variable with a similar name exists: `username`" (already partial).
    - [ ] Missing return: "help: add `-> ReturnType` to function signature or remove `return`".
    - [ ] Optional misuse: "help: use `value?` to unwrap or `value ?? default` for fallback".

- [ ] **"Did You Mean?" Enhancements**
    - [ ] Improve Levenshtein distance threshold based on identifier length.
    - [ ] Consider scope (prefer suggestions from current/parent scope over globals).
    - [ ] Suggest methods when calling undefined function on a typed value.
    - [ ] Suggest imports when using undefined types that exist in stdlib.

### Error Codes & Documentation

- [ ] **Error Code System**
    - [ ] Assign unique codes to each error type (e.g., `E001`, `E002`).
    - [ ] Display code in error output: `error[E042]: type mismatch`.
    - [ ] Create `--explain E042` flag to show detailed explanation.
    - [ ] Generate error index documentation (for website/docs).

- [ ] **Detailed Explanations**
    - [ ] Write extended explanations for each error code.
    - [ ] Include: why this is an error, common causes, examples of fixes.
    - [ ] Link to relevant language documentation.

### Error Aggregation & Prioritization

- [ ] **Smart Error Limiting**
    - [ ] Limit to N errors per file (default 10), with "and N more errors" summary.
    - [ ] Prioritize errors: syntax errors first, then semantic, then warnings.
    - [ ] Deduplicate cascading errors (e.g., undefined variable used 10 times → 1 error).

- [ ] **Related Error Grouping**
    - [ ] Group errors by location (same function/class).
    - [ ] Identify and suppress cascading errors caused by earlier errors.
    - [ ] "Fix this first" hints for errors that likely cause others.

### Example Target Output

```
error[E023]: type mismatch in argument
  --> src/main.rg:15:12
   |
14 |   user = find_user(name)
15 |   greet(user.age)
   |         ^^^^^^^^
   |         |
   |         expected String, found Int
   |
note: function `greet` defined here
  --> src/main.rg:3:1
   |
 3 | def greet(name : String)
   | ^^^^^^^^^^^^^^^^^^^^^^^^
   |
help: convert Int to String with `.to_s`
   |
15 |   greet(user.age.to_s)
   |                 +++++
```

## Phase 2: Architecture & Maintainability

- [x] **Visitor Pattern** - DECIDED: Keep current approach ✅
    - The type-switch dispatch pattern in `genStatement` and `genExpr` is idiomatic Go and readable.
    - Codegen is already well-organized: `statements.go`, `expressions.go`, `declarations.go`.
    - Added doc comments and logical groupings to switch statements for maintainability.
    - *Revisit if:* We need multiple code generation backends (e.g., LLVM) or cross-cutting concerns.

- [ ] **Dependency Graph & Build System**
    - [ ] Analyze imports to build a dependency graph of `.rg` files.
    - [ ] Support parallel compilation of independent modules.
    - [ ] Implement smarter incremental builds.

## Phase 3: Code Generation Optimizations

- [x] **Monomorphization / Generics** ✅
    - [x] Utilize the Typed AST from Phase 1.
    - [x] Generate native Go types (`[]int`, `map[string]int`) instead of `[]any` / `map[any]any` for literals.
    - [x] Extended `TypeInfo` interface with `GetGoType()` method for Go type string generation.
    - [x] Added `GoType()` method to `semantic.Type` for centralized type-to-Go conversion.
    - [ ] Further optimization: propagate types through variable assignments and function returns.

- [ ] **Zero-Value Fixes**
    - [ ] Ensure `zeroValue` generation logic correctly handles interfaces and custom types to avoid runtime panics or nil pointer dereferences.

- [ ] **Module/Mixin Overhaul**
    - [ ] Move away from "copy-paste" flattening of modules into classes.
    - [ ] Investigate compiling Modules to embedded Go structs.
    - [ ] Implement robust conflict resolution (Diamond Problem) during the semantic phase.

- [x] **Direct Comparisons & Math Operations** ✅
    - [x] Math operations already use direct `x + y` (not runtime calls).
    - [x] Equality comparisons (`==`, `!=`) now use direct comparison for primitive types when type info is available.
    - [x] Added `TypeInfo` interface to pass type information from semantic analysis to codegen.
    - [x] Falls back to `runtime.Equal` for classes/unknown types (custom equality support).

## Phase 4: Debugging & Tooling

- [ ] **LSP Server (Language Server Protocol)**
    - [ ] Leverage the Semantic Analysis phase to provide:
        - [ ] Hover types.
        - [ ] Go-to-definition.
        - [ ] Autocomplete.

- [ ] **Enhanced Source Maps**
    - [ ] Improve `//line` directive emission.
    - [ ] Ensure panics in generated Go code point precisely to the Rugby source line.

- [ ] **Linter / Static Analysis**
    - [ ] Built upon the Semantic Analysis phase.
    - [ ] Detect unused variables, unreachable code, and style violations.
