# Rugby Development Roadmap

Features in implementation order, building incrementally.

**Recent spec updates (see spec.md):**
- **Strict Conditionals (5.2):** Only `Bool` allowed in `if`/`while`. No implicit truthiness.
- **Unified Optionals (4.4):** `T?` always maps to `Option[T]` struct (storage) or `(T, bool)` (return).
- **Explicit Typing (6.1):** Function parameters must be explicitly typed (except `any`).
- **Strict Field Inference (7.2):** Fields inferred from `initialize` or declared explicitly; new fields outside `initialize` are errors.
- **Pure Blocks (5.4):** No `break`/`next` in blocks. `return` exits block only.
- **Strict Calls (6.5):** Parentheses required for method calls (except properties).
- **Interface Evolution (9):** `implements`, `any`, interface inheritance, `is_a?`, `as`.

## Phase 16: Interface System Evolution (In Progress)
**Goal:** Align interfaces with the refined class model and provide safe runtime polymorphism.
- [ ] **Interface Inheritance:** Support `interface IO < Reader, Writer` syntax and Go embedding.
- [ ] **`implements` Keyword:** Support `class User implements Speaker` for compile-time conformance checks.
- [ ] **`any` Keyword:** Map `any` to Go's `any` (empty interface).
- [ ] **Runtime Casting:**
  - [ ] Implement `is_a?(Interface)` → Go type assertion.
  - [ ] Implement `as(Interface)` and `as?(Interface)` casting.
- [ ] **Standard Interface Mapping:** Ensure `to_s` satisfies `fmt.Stringer` and `message` satisfies `error`.

## Phase 17: Crispness Polish (Strictness & Safety)
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

### 17.2 Strict Semantics & Codegen
- [ ] **Strict Conditionals:**
  - Update Codegen to validate that `if`/`while` conditions evaluate to `Bool`.
  - Remove implicit truthiness logic.
- [ ] **Unified Optionals:**
  - Add `runtime.Option[T]` struct.
  - Update Codegen to use `Option[T]` for `T?` fields and variables.
  - Update `runtime/optional.go` helpers.
  - Implement `if let` pattern in Parser and Codegen.

### 17.3 Type System Refinement
- [ ] **Case vs Type Switch:**
  - Introduce `case_type` (or `case type`) syntax for type switching.
  - Restrict standard `case` to value matching (`==`).

## Phase 18: Standard Library Polish
- [ ] **String Interpolation:** Ensure all interpolations compile to `fmt.Sprintf` (or `String()` calls).
- [ ] **Range Constraints:** Restrict `Range` to `Int` only.

---

## Recently Completed

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
