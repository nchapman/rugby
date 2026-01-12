# Rugby Development Roadmap

Features in implementation order, building incrementally.

**Recent spec updates (see spec.md):**
- **Strict Conditionals (5.2):** Only `Bool` allowed in `if`/`while`. No implicit truthiness.
- **Unified Optionals (4.4):** `T?` with operators (`??`, `&.`) and methods (`ok?`, `unwrap!`, `map`, etc.).
- **The `nil` Keyword (4.5):** Only `T?` and `error` can be `nil`. Value/reference types require `T?` for nullability.
- **Explicit Typing (6.1):** Function parameters must be explicitly typed (except `any`).
- **Strict Field Inference (7.2):** Fields inferred from `initialize` or declared explicitly; new fields outside `initialize` are errors.
- **Pure Blocks (5.4):** No `break`/`next` in blocks. `return` exits block only.
- **Strict Calls (6.5):** Parentheses required for method calls (except properties).
- **Interface Evolution (9):** `implements`, `any`, interface inheritance, `is_a?`, `as` (returns `T?`).
- **Error Handling (15):** `error` type, `error()` kernel function, postfix `!` propagation, `rescue` keyword, `panic` for unrecoverable errors.
- **Strict `?` Suffix (10.3):** Methods ending in `?` must return `Bool`. No `to_i?`/`as?` patterns.

## Phase 16: Interface System Evolution ✓
**Goal:** Align interfaces with the refined class model and provide safe runtime polymorphism.
- [x] **Interface Inheritance:** Support `interface IO < Reader, Writer` syntax and Go embedding.
- [x] **`implements` Keyword:** Support `class User implements Speaker` for compile-time conformance checks.
- [x] **`any` Keyword:** Map `any` to Go's `any` (empty interface).
- [x] **Runtime Casting:**
  - [x] Implement `is_a?(Interface)` → Go type assertion returning `Bool`.
  - [x] Implement `as(Interface)` → returns `Interface?` (optional), use with `if let`, `??`, `.unwrap!`.
- [x] **Standard Interface Mapping:** Ensure `to_s` satisfies `fmt.Stringer` and `message` satisfies `error`.

## Phase 17: Crispness Polish (Strictness & Safety) ✓
**Goal:** Reduce ambiguity and enforce deterministic behavior.

### 17.1 Strict Syntax & Parsing ✓
- [x] **Strict Method Calls:** Parser requires parentheses for all method calls.
  - Exception: Property accessors (getters/setters).
- [x] **Pure Blocks:** Parser rejects `break` and `next` keywords inside blocks.
- [x] **Explicit Parameters:** Parser requires type annotations for all function parameters (use `: any` for untyped).
- [x] **Strict Field Inference:**
  - Parser parses field declarations in class body (`@x : Type`).
  - Fields can be introduced via: explicit declaration, parameter promotion, or first assignment in `initialize`.
  - New fields outside `initialize` are compile-time errors.

### 17.2 Strict Semantics & Codegen ✓
- [x] **Strict Conditionals:**
  - Update Codegen to validate that `if`/`while` conditions evaluate to `Bool`.
  - Remove implicit truthiness logic.
- [x] **Unified Optionals:**
  - Add `runtime.Option[T]` struct.
  - Implement `if let` pattern in Parser and Codegen.
  - Implement tuple unpacking: `val, ok = optional_expr` → `(T, Bool)`.
  - Implement optional operators:
    - `??` (nil coalescing): `T? ?? T → T`
    - `&.` (safe navigation): `T?&.method → R?`
  - Implement optional methods:
    - `ok?` / `present?` → Bool
    - `nil?` / `absent?` → Bool
    - `unwrap!` → T (dereference)

### 17.3 Type System Refinement ✓
- [x] **Case vs Type Switch:**
  - Introduced `case_type` syntax for type switching.
  - Standard `case` restricted to value matching (`==`).

### 17.4 Strict `?` Suffix ✓
- [x] **Predicate Enforcement:** Methods ending in `?` must return `Bool`.
  - Codegen validates return type of `?`-suffixed methods.
  - Compile error if `def foo? -> Int` or similar.
- [x] **Remove Optional-Returning `?` Methods:**
  - `to_i` returns `(Int, error)` instead of `to_i?` returning `Int?`.
  - `to_f` returns `(Float, error)` instead of `to_f?` returning `Float?`.
  - Updated runtime string conversion functions accordingly.

## Phase 18: Standard Library Polish
- [ ] **String Interpolation:** Ensure all interpolations compile to `fmt.Sprintf` (or `String()` calls).
- [ ] **Range Constraints:** Restrict `Range` to `Int` only.

## Phase 19: Error Handling ✓
**Goal:** Implement Go-style error handling with ergonomic syntax sugar.

### 19.1 Core Error Support ✓
- [x] **`error` Type:** Map Rugby `error` to Go's `error` interface.
- [x] **Error Returns:** Support `-> error` and `-> (T, error)` return signatures.
- [x] **`nil` as No Error:** Allow `nil` as valid error value.
- [x] **`error()` Kernel Function:** Create error values with `error("message")` - maps to `runtime.Error()`.

### 19.2 Postfix Bang Operator (`!`) ✓
- [x] **Lexer:** Add `BANG` token recognition for postfix `!` on call expressions.
- [x] **Parser:** Parse `call_expr!` as error propagation.
  - Only valid on call expressions with parentheses.
  - `f!` is a method name, `f()!` is propagation.
- [x] **Codegen:** Lower `call!` to Go error check pattern:
  ```go
  result, err := call()
  if err != nil { return <zero-values>, err }
  ```
- [x] **Enclosing Function Check:** Compile error if `!` used in function not returning `error`.
- [x] **Chained Calls:** Support `a()!.b()!.c()!` with sequential unwrapping.

### 19.3 Script Mode / Main ✓
- [x] **`runtime.Fatal`:** Implement `Fatal(err)` that prints to stderr and exits with code 1.
- [x] **Main Lowering:** In `def main` or top-level scripts, `call!` lowers to `runtime.Fatal(err)`.

### 19.4 Rescue Keyword ✓
- [x] **Lexer:** Add `RESCUE` and `FATARROW` (`=>`) tokens.
- [x] **Parser:** Parse three forms:
  - Inline: `call() rescue default_expr`
  - Block: `call() rescue do ... end`
  - Block with binding: `call() rescue => err do ... end`
- [x] **Codegen:**
  - Inline: `if err != nil { result = default }`
  - Block: `if err != nil { <block statements>; result = <last expr> }`
  - Binding: `if err != nil { err := err; <block> }`
- [x] **Mutual Exclusion:** Compile error if `!` and `rescue` both appear on same call.

### 19.5 Panic (Unrecoverable Errors) ✓
- [x] **Lexer:** Add `PANIC` token (changed from `RAISE` for clarity).
- [x] **Parser:** Parse `panic expr` as panic statement.
- [x] **Codegen:** Lower to `panic(expr)` or `panic(fmt.Sprintf(...))` for interpolated strings.
- **Note:** Renamed from `raise` to `panic` to avoid confusion with Ruby's catchable exceptions.

### 19.6 Error Utilities
- [ ] **`error_is?(err, target)`:** Compile to `errors.Is(err, target)`, returns `Bool`.
- [ ] **`error_as(err, Type)`:** Compile to `errors.As` pattern, returns `Type?` (optional).
  - Use with `if let` for unwrapping.
  - Auto-import `"errors"` package when these functions are used.

### 19.7 Runtime Updates ✓
- [x] Add `runtime.Fatal(err error)` function.
- [x] Update `runtime.StringToInt` to return `(int, error)` instead of optional.
- [x] Update `runtime.StringToFloat` to return `(float64, error)` instead of optional.

---

## Recently Completed

### Type System Refinement & Strict `?` Suffix (Phase 17.3 & 17.4)
- **`case_type` syntax:** Type switching for runtime type matching on `any` values
  - Parses `case_type x when String ... when Int ... end`
  - Generates Go type switch: `switch x := x.(type) { case string: ... }`
  - Auto-narrows subject variable inside each branch
- **Strict `?` suffix:** Methods ending in `?` must return `Bool`
  - Codegen validates return type and errors if not `Bool`
  - Example error: `method 'valid?' ending in '?' must return Bool`
- **String conversions return error:**
  - `to_i` now returns `(Int, error)` - use with `!` or `rescue`
  - `to_f` now returns `(Float, error)` - use with `!` or `rescue`
  - Removed `to_i?` and `to_f?` optional patterns

### Interface System Evolution (Phase 16)
Complete interface support with inheritance, runtime polymorphism, and standard interface mapping:
- **Interface Inheritance**: `interface IO < Reader, Writer` compiles to Go interface embedding
- **`implements` Keyword**: `class User implements Speaker` generates compile-time conformance checks via `var _ Speaker = (*User)(nil)`
- **`any` Keyword**: Maps to Go's `any` (empty interface) for generic parameters
- **Runtime Casting**:
  - `is_a?(Type)`: Type predicate returning `Bool`, compiles to Go type assertion with `_, ok := obj.(Type)`
  - `as(Type)`: Type cast returning `(Type, bool)`, compiles to Go type assertion with value
- **Standard Interface Mapping**:
  - `to_s` → `String()` (satisfies `fmt.Stringer`)
  - `message` → `Error()` (satisfies `error` interface)
- Parser updated to accept `any` keyword in type positions and `as` in selector expressions
- Comprehensive tests added in `codegen/interface_features_test.go`

### Error Handling (Phase 19)
Complete Go-style error handling with Rugby-idiomatic syntax:
- **`error` type:** Maps to Go's `error` interface for function returns
- **`error()` kernel function:** Creates error values (`error("message")`) - no imports needed
- **Postfix `!` operator:** Error propagation with context-aware behavior
  - In error-returning functions: early return with zero values
  - In `main`/scripts: calls `runtime.Fatal()` and exits
  - Supports chained calls: `a()!.b()!.c()!`
- **`rescue` keyword:** Inline error recovery with three forms:
  - Inline default: `call() rescue default_value`
  - Block: `call() rescue do ... end`
  - Block with error binding: `call() rescue => err do ... end`
- **`panic` statement:** For unrecoverable programmer errors (not `raise`)
  - Renamed from `raise` to avoid confusion with Ruby's catchable exceptions
  - Compiles directly to Go's `panic()`
- **`runtime.Fatal()`:** Prints error to stderr and exits with status 1

### Strict Semantics & Codegen (Phase 17.2)
- **Strict Conditionals:** Only `Bool` allowed in `if`/`while` conditions. Removed implicit truthiness.
  - Codegen now validates condition types and returns errors for non-Bool conditions
  - Nil comparisons (`x != nil`) still allowed as they return `Bool`
- **Unified Optionals:** Complete optional type support with operators and methods:
  - `runtime.Option[T]` generic struct for field storage
  - `if let` pattern for optional unwrapping in conditions
  - Tuple unpacking with `val, ok = expr` syntax
  - `??` (nil coalescing): Returns value if present, otherwise default
  - `&.` (safe navigation): Safely calls method on optional, returns nil if absent
  - Optional methods: `ok?`/`present?`, `nil?`/`absent?`, `unwrap!`

### Strict Field Inference (Phase 17.1)
- Parser now supports explicit field declarations (`@field : Type` in class body)
- Implemented parameter promotion syntax (`def initialize(@field : Type)`)
- Fields inferred from first assignment in `initialize` method
- Validation: methods other than `initialize` cannot introduce new instance variables
- Formatter updated to display field declarations with proper indentation

### Parser Restructure
Split the monolithic `parser.go` (2,569 lines) into focused modules:
- `parser.go` - Core infrastructure (238 lines)
- `declarations.go` - Function, class, interface declarations
- `statements.go` - Control flow statements
- `expressions.go` - Pratt parser for expressions
- `blocks.go` - Block parsing (do/end, braces)
- `literals.go` - Literal parsing (int, float, string, array, map)
- `errors.go` - Structured ParseError with hints
- `precedence.go` - Operator precedence definitions
- `testing.go` - Test DSL parsing (describe, it, test, table)

### Builder Enhancement
- Added `isInRugbyRepo()` to auto-detect development environment
- Auto-injects `replace` directive for local runtime during development
