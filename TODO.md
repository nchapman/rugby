# Rugby Development Roadmap

Features in implementation order, building incrementally.

**Recent spec updates (see spec.md):**
- **Strict Conditionals (5.2):** Only `Bool` allowed in `if`/`while`. No implicit truthiness.
- **Unified Optionals (4.4):** `T?` always maps to `Option[T]` struct (storage) or `(T, bool)` (return).
- **Explicit Typing (6.1):** Function parameters must be explicitly typed (except `any`).
- **Explicit Fields (7.1):** All instance variables must be declared in class body.
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
- [ ] **Explicit Fields:**
  - Update Parser to parse field declarations in class body (`@x : Int`).
  - Update Parser/Builder to require all used `@ivars` to be declared.
  - Remove implicit field inference from `initialize`.

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
