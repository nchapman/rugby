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

## Phase 1.5: Testing & Development UX (Critical for Refactoring)

Testing generated string output is fragile, and single-error stopping slows down development. We need behavioral tests and better error reporting to refactor `codegen` with confidence.

- [x] **Integration Test Harness** ✅
    - [x] Harness exists in `examples/examples_test.go` - runs `rugby build`, executes binaries, asserts `stdout`/`stderr`/`exit_code`.
    - [x] Covers hello, arith, fizzbuzz, case, ranges, symbols, errors, json, http examples.

- [x] **Parser Error Recovery** ✅
    - [x] Parser already reports multiple errors per file (collects errors in slice, continues parsing).
    - [x] Added `synchronize()` and `syncToEnd()` utilities for future recovery improvements.
    - [x] Test added verifying multiple errors reported in one parse pass.

## Phase 2: Architecture & Maintainability

- [ ] **Visitor Pattern**
    - [ ] Define a `Visitor` interface for AST nodes (`Accept(v Visitor)`).
    - [ ] Refactor `codegen` to implement the `Visitor` interface.
    - *Note:* Prioritize this **after** the semantic pass is working. Avoid over-engineering; simpler switch statements are acceptable if they remain readable.

- [ ] **Dependency Graph & Build System**
    - [ ] Analyze imports to build a dependency graph of `.rg` files.
    - [ ] Support parallel compilation of independent modules.
    - [ ] Implement smarter incremental builds.

## Phase 3: Code Generation Optimizations

- [ ] **Monomorphization / Generics**
    - [ ] Utilize the Typed AST from Phase 1.
    - [ ] Generate native Go types (`[]int`, `map[string]User`) instead of `[]any` / `map[any]any`.
    - [ ] Reduce runtime reflection and interface conversion overhead.

- [ ] **Zero-Value Fixes**
    - [ ] Ensure `zeroValue` generation logic correctly handles interfaces and custom types to avoid runtime panics or nil pointer dereferences.

- [ ] **Module/Mixin Overhaul**
    - [ ] Move away from "copy-paste" flattening of modules into classes.
    - [ ] Investigate compiling Modules to embedded Go structs.
    - [ ] Implement robust conflict resolution (Diamond Problem) during the semantic phase.

- [ ] **Direct Math Operations**
    - [ ] Replace `runtime.Add(x, y)` with `x + y` when types are known to be primitive `Int`/`Float`.
    - [ ] Inline simple comparisons.

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
