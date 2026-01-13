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

## Phase 1.75: Error Experience ✅

Compiler errors display with source context, colors, and precise locations.

- [x] Source context: 1-3 lines of code with line numbers
- [x] Colored output with TTY detection and `NO_COLOR` support
- [x] `--color=always|never|auto` flag
- [x] Underlines (`^~~~`) pointing to error location
- [x] "Did you mean?" suggestions for typos

Example output:
```
error: undefined: 'username' (did you mean 'user_name'?)
  --> src/main.rg:15:5
   |
14 |   user = find_user(name)
15 |   puts username
   |        ^^^^^^^^
16 | end
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
    - [x] Type propagation: `inferTypeFromExpr()` now uses semantic type info for variables assigned from function calls.

- [x] **Zero-Value Fixes** ✅
    - [x] Added `interfaces` map to Generator to track declared interface types.
    - [x] `zeroValue()` now returns `nil` for custom interfaces (previously generated invalid `InterfaceName{}`).
    - [x] Added safeguard for qualified Go types (e.g., `io.Reader`) to return `nil`.

- [x] **Module/Mixin Conflict Detection** ✅
    - [x] Detect duplicate methods/accessors when multiple modules are included.
    - [x] Class methods/accessors intentionally override module versions (no error).
    - [x] Two modules defining same method/accessor produces clear error message.
    - [x] Added `Errors()` method to Generator for accessing collected errors.

- [ ] **Module/Mixin Architecture** (Future)
    - [ ] Move away from "copy-paste" flattening of modules into classes.
    - [ ] Investigate compiling Modules to embedded Go structs.

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

- [x] **Enhanced Source Maps** ✅
    - [x] Added `Line` fields to all AST statement types (OrAssignStmt, CompoundAssignStmt, MultiAssignStmt, BreakStmt, NextStmt, DeferStmt).
    - [x] Updated `getStatementLine()` to cover all statement types (declarations, assignments, control flow, jump statements, concurrency, testing).
    - [x] Parser captures line numbers at statement start for accurate source mapping.
    - [x] Added tests for line directive emission.
    - [x] Panics in generated Go code now point precisely to the Rugby source line.

- [ ] **Linter / Static Analysis**
    - [ ] Built upon the Semantic Analysis phase.
    - [ ] Add `Used` field to Symbol for tracking variable usage.
    - [ ] Detect unused variables.
    - [ ] Detect unreachable code after return/break/next.
    - [ ] Add linter CLI integration (`rugby lint`).
