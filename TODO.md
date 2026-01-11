# Rugby MVP TODO

Features in implementation order, building incrementally.

**Recent spec updates (see spec.md):**
- **Runtime package (11):** Full Go runtime with Ruby-like stdlib ergonomics
- **Loop semantics (5.3-5.4):** Clear distinction between imperative loops (`for`/`while` with `break`/`next`/`return` exits function) and functional blocks (`each`/`map` where `return` exits block only)
- **Range type (4.2.1):** First-class Range values with `..`/`...` syntax, methods, and loop integration (Phase 10)
- **Special methods (7.8):** `def to_s` → `String() string`; `def ==(other)` → `Equal()` method; `==` operator uses `runtime.Equal` with dispatch logic
- **self keyword (7.9):** `self` refers to receiver, enables method chaining
- Return statements (6.3): explicit + implicit returns
- Instance variables (7.2.1): types inferred from `initialize`
- Optionals (4.4): `T?` restricted to value types only
- Interface methods (8.1): implicitly exported
- String interpolation (3.3): any expression allowed
- Map iteration (5.3): `map.each do |k, v|`

## Phase 1: Core Expressions
- [x] Integer literals and arithmetic (`+`, `-`, `*`, `/`, `%`)
- [x] Float literals
- [x] Boolean literals (`true`, `false`)
- [x] Comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`)
- [x] Boolean operators (`and`, `or`, `not` / `&&`, `||`, `!`)
- [x] Parenthesized expressions

## Phase 2: Variables & Control Flow
- [x] Variable assignment (`x = 5`)
- [x] Variable references
- [x] Conditionals (`if`/`elsif`/`else`/`end`)
- [x] While loops (`while cond ... end`)
- [x] For loops (`for item in items ... end`)
- [x] Loop control: `break` (exit loop), `next` (continue to next iteration)
- [x] Return statements

## Phase 3: Functions
- [x] Function parameters (`def add(a, b)`)
- [x] Explicit return types (`def add(a, b) -> Int`)
- [x] Multiple return values (`-> (Int, Bool)`)
- [x] Function calls with arguments (`add(1, 2)`)

## Phase 4: Go Interop
- [x] Method calls with dot syntax (`http.Get(url)`)
- [x] Import aliases (`import encoding/json as json`)
- [x] snake_case to CamelCase mapping (`read_all` -> `ReadAll`)
- [x] Defer (`defer f.Close`)

**Note on snake_case transformation (see spec.md Section 9.4):**
Currently, snake_case → PascalCase only applies to selector calls (Go interop).
Local function names are passed through as-is. Full behavior once `pub` is implemented:
- Selector calls (`io.read_all`) → PascalCase (`io.ReadAll`) - Go exports are always uppercase
- `pub def parse_json` → `func ParseJSON()` - exported
- `def parse_json` → `func parseJSON()` - internal (camelCase)
- Acronym handling (spec 9.5): `id` → `ID`, `http` → `HTTP`, etc.
- Name collision detection (spec 9.6): compile-time errors for collisions
- Reserved word escaping (spec 9.7): suffix `_` for internal, error for `pub`
This requires `pub` keyword support (Phase 8) to be fully correct.

## Phase 5: Collections & Blocks
- [x] Array literals (`[1, 2, 3]`)
- [x] Array indexing (`arr[0]`)
- [x] Map literals (`{"a" => 1}`)
- [x] Map access (`m["key"]`)
- [x] `each` blocks (`arr.each do |x| ... end`)
- [x] `each_with_index`
- [x] `map` blocks (`arr.map do |x| ... end`)
- [x] **Generic block architecture** (blocks now method-agnostic)
  - [x] Add `BlockExpr` AST node (generic `do |params| ... end`)
  - [x] Update `CallExpr` to support optional block argument
  - [x] Refactor parser: parse all blocks generically
  - [x] Update codegen: blocks use runtime calls (`runtime.Each`, `runtime.Times`, etc.) so `return` exits block only
  - [x] Proper variable scope tracking in blocks
  - [x] Duplicate block parameter detection
  - [x] Nested blocks support

## Phase 5b: Runtime Package
Create `runtime/` Go package providing Ruby-like stdlib ergonomics (see spec.md Section 11).

### Setup
- [x] Create `runtime/` directory structure
- [x] Set up Go module for runtime package
- [x] Add runtime import generation to codegen (auto-import when used)
- [x] Support `?` and `!` suffixes in method names (lexer)

### Array methods (`runtime/array.go`)
- [x] `Each(slice, func(elem))` - iterate with block semantics
- [x] `EachWithIndex(slice, func(elem, index))` - iterate with index
- [x] `Select[T]([]T, func(T) bool) []T` - filter elements
- [x] `Reject[T]([]T, func(T) bool) []T` - inverse filter
- [x] `Map[T,R]([]T, func(T) R) []R` - transform elements (replace IIFE)
- [x] `Reduce[T,R]([]T, R, func(R,T) R) R` - fold/accumulate
- [x] `Find[T]([]T, func(T) bool) (T, bool)` - first match
- [x] `Any[T]([]T, func(T) bool) bool` - any match?
- [x] `All[T]([]T, func(T) bool) bool` - all match?
- [x] `None[T]([]T, func(T) bool) bool` - no match?
- [x] `Contains[T comparable]([]T, T) bool` - includes value?
- [x] `First[T]([]T) (T, bool)` - first element
- [x] `Last[T]([]T) (T, bool)` - last element
- [x] `Reverse[T]([]T)` - in-place reverse
- [x] `Reversed[T]([]T) []T` - return reversed copy
- [x] `Sum` / `Min` / `Max` for numeric slices

### Map methods (`runtime/map.go`)
- [x] `Keys[K,V](map[K]V) []K`
- [x] `Values[K,V](map[K]V) []V`
- [x] `Merge[K,V](map[K]V, map[K]V) map[K]V`
- [x] `MapSelect[K,V](map[K]V, func(K,V) bool) map[K]V`
- [x] `MapReject[K,V](map[K]V, func(K,V) bool) map[K]V`
- [x] `Fetch[K,V](map[K]V, K, V) V` - get with default

### String methods (`runtime/string.go`)
- [x] `Chars(string) []string` - split into characters
- [x] `CharLength(string) int` - rune count
- [x] `StringReverse(string) string`
- [x] `StringToInt(string) (int, bool)` - safe parse
- [x] `StringToFloat(string) (float64, bool)`
- [x] `MustAtoi(string) int` - panic on failure

### Integer methods (`runtime/int.go`)
- [x] `Times(n, func(i))` - iterate n times
- [x] `Upto(start, end, func(i))` - iterate from start to end inclusive
- [x] `Downto(start, end, func(i))` - iterate from start down to end inclusive
- [x] `Abs(int) int`
- [x] `Clamp(int, int, int) int`

### Global/Kernel functions (`runtime/io.go`)
- [x] `Puts(args...)` - print with newline (like Ruby puts)
- [x] `Print(args...)` - print without newline
- [x] `P(args...)` - debug print with %#v formatting
- [x] `Gets()` - read line from stdin
- [x] `GetsWithPrompt(prompt)` - print prompt, read line
- [x] `Exit(code)` - terminate program
- [x] `Sleep(seconds)` - pause execution (float seconds)
- [x] `SleepMs(ms)` - pause execution (int milliseconds)
- [x] `RandInt(n)` - random int [0, n)
- [x] `RandFloat()` - random float [0.0, 1.0)
- [x] `RandRange(min, max)` - random int [min, max]

### Codegen updates
- [x] Update `map` to use `runtime.Map()` instead of IIFE
- [x] Add `select`/`filter` block codegen → `runtime.Select()`
- [x] Add `reject` block codegen → `runtime.Reject()`
- [x] Add `reduce` block codegen → `runtime.Reduce()`
- [x] Add `find` block codegen → `runtime.Find()`
- [x] Add kernel functions codegen (`puts`, `print`, `p`, `gets`, `exit`, `sleep`, `rand`)
- [x] Auto-import `rugby/runtime` when runtime functions are used
- [x] Add predicate methods: `any?`, `all?`, `none?`
- [x] Refactor to table-driven mappings (removed 146 lines of duplicate code)
- [x] Add `times`, `upto`, `downto` integer iteration
- [x] Add `runtime.Equal(a, b)` - dispatch logic:
  1. If `a` has `Equal(interface{}) bool` method, call it
  2. Else if slices/maps, deep comparison
  3. Else use Go `==`

## Phase 5c: Runtime Extensions (Extended Stdlib)
**Goal:** Fill in gaps to match common Ruby stdlib usage.

### Array (`runtime/array.go`)
- [x] `Join(slice, separator) string`
- [x] `Flatten(slice) []interface{}` (recursive or shallow?)
- [x] `Uniq[T](slice) []T`
- [x] `Sort[T](slice) []T` (requires Orderable constraint or comparator)
- [x] `Shuffle[T](slice) []T`
- [x] `Sample[T](slice) T`
- [x] `FirstN[T](slice, n) []T`
- [x] `LastN[T](slice, n) []T`
- [x] `Rotate[T](slice, n) []T`

### Map (`runtime/map.go`)
- [x] `Delete(map, key)`
- [x] `HasKey(map, key) bool`
- [x] `Clear(map)`
- [x] `Invert(map) map` (value to key)

### String (`runtime/string.go`)
- [x] `Split(str, sep) []string`
- [x] `Strip(str) string` / `Lstrip` / `Rstrip`
- [x] `Upcase(str) string` / `Downcase(str) string`
- [x] `Capitalize(str) string`
- [x] `Contains(str, substr) bool`
- [x] `Replace(str, old, new) string`

### Integer/Math (`runtime/int.go` / `runtime/math.go`)
- [x] `Even(int) bool`
- [x] `Odd(int) bool`
- [x] `Sqrt(float) float`
- [x] `Pow(base, exp) float`

### Codegen
- [x] Update `codegen/codegen.go` to map new methods to runtime functions

## Phase 6: Classes
- [x] Class definition (`class User ... end`)
- [x] Instance variables (`@name`) - types inferred from `initialize`
- [x] Constructor (`def initialize`) - generates `New*` functions
- [x] `Class.new(...)` syntax
- [x] Methods with pointer receivers (for mutation support)
- [x] Embedding via `<` (`class Service < Logger`) - single parent only for MVP
- [x] Pointer receiver methods (`def mutate!`) - `!` suffix stripped in Go output
- [x] Methods are lowercase (private) by default - `pub` enables uppercase exports
- [x] `self` keyword → compiles to receiver variable (e.g., `u` in `func (u *User)`)
- [x] Special method `to_s` → `String() string` (satisfies `fmt.Stringer`)
- [x] Custom equality: `def ==(other)` → `Equal(other interface{}) bool`
- [x] Equality operator: `==` compiles to `runtime.Equal(a, b)` for non-primitives

## Phase 7: Type Annotations
**Goal:** Add explicit type annotations following Go's philosophy.

### Lexer
- [x] Add COLON token (`:`)

### Parser
- [x] Parse variable type annotations: `x : Int = 5`
- [x] Parse parameter type annotations: `def foo(x : String, y : Int)`
- [x] Update Param AST node to include optional Type field
- [x] Update AssignStmt AST node to include optional Type field

### Codegen
- [x] Generate typed variable declarations: `x : Int = 5` → `var x int = 5`
- [x] Generate typed function parameters: `def foo(x : Int)` → `func foo(x int)`
- [x] Infer instance variable types from initialize parameter types
  - Previously: `@name` → `name interface{}`
  - Now: `def initialize(name : String)` + `@name = name` → `name string`
- [x] Keep `interface{}` default when no type annotation provided

### Type mappings (already exist in mapType function)
- Rugby → Go: Int→int, String→string, Bool→bool, Float→float64, etc.

### Testing
- [x] Lexer tests for COLON token
- [x] Parser tests for type annotations
- [x] Codegen tests for typed variables, parameters, and instance variables
- [x] Verify untyped code still works (backward compatibility)

### Deferred to later phases:
- Optional types (`T?`) - complex, needs careful design
- Generic type parameters (`Array[T]`, `Map[K,V]`)
- `nil` literal handling
- Type inference engine

## Phase 8: Interfaces & Visibility
- [x] Interface definitions (`interface Speaker`) - methods implicitly exported
- [x] Structural conformance checking (deferred to Go compiler)
- [x] `pub` keyword for exports - enables proper name casing
- [x] Acronym handling: `id` → `ID`, `url` → `URL`, `http` → `HTTP`, etc. (spec 9.5)
- [x] Validation: `pub def` in non-pub class is compile error (spec 9.2)

## Phase 9: Strings & Polish
- [x] String interpolation (`"hi #{name}"`) - any expression, uses `fmt.Sprintf`
- [x] Comments in all positions (trailing, inline) - already worked
- [x] Better error messages with `file:line:column` format
- [x] Multi-file compilation (`rugby file1.rg file2.rg` or `rugby directory/`)

## Phase 10: Range Type
First-class Range values (see spec.md Section 4.2.1).

### Lexer/Parser
- [x] `..` token (inclusive range)
- [x] `...` token (exclusive range)
- [x] `RangeLit` AST node with `Start`, `End`, `Exclusive` fields
- [x] Parse range as infix operator (low precedence)

### Runtime (`runtime/range.go`)
- [x] `Range` struct: `Start`, `End int`, `Exclusive bool`
- [x] `RangeEach(r, fn)` - iterate with block
- [x] `RangeContains(r, n)` - membership test
- [x] `RangeToArray(r)` - materialize to `[]int`
- [x] `RangeSize(r)` - element count

### Codegen
- [x] Range literals → `runtime.Range{Start: x, End: y, Exclusive: bool}`
- [x] `for i in range` → C-style for loop: `for i := start; i <= end; i++`
- [x] Range method calls → runtime function calls

## Phase 11: High-Value Features

### Multiple Embedding
- [x] Support `class Foo < Bar, Baz` syntax (spec 7.6)
- [x] AST: `ClassDecl.Embeds []string` (was `Parent string`)
- [x] Parser: Comma-separated list of embedded types
- [x] Codegen: Emit multiple embedded fields in struct

### `||=` Operator (spec 5.1)
- [x] Token: `ORASSIGN` for `||=`
- [x] Lexer: Handle `||=` while preserving `||` as two PIPE tokens for blocks
- [x] AST: `OrAssignStmt` and `InstanceVarOrAssign` nodes
- [x] Parser: Detect `||=` after identifiers and instance variables
- [x] Codegen: First use → `:=`, subsequent use → nil-check pattern
- [x] Field extraction: Handle `||=` in initialize methods

**Note:** Per spec, `||=` is not applicable for non-nullable value types (Int, String, etc.) - these will fail at Go compile time, which is correct behavior.

### Optional Types `T?` (spec 4.4) - Phase A
- [x] Lexer: `?` suffix consumed as part of type identifier (e.g., `Int?`)
- [x] Parser: `parseTypeName()` handles optional types
- [x] Codegen: `mapType()` handles optionals:
  - Value types (`Int?`, `String?`) → `runtime.OptionalT`
  - Reference types (`User?`) → `*User` pointer
  - Slices/maps already nullable, passed through
- [x] Runtime: `optional.go` with OptionalInt, OptionalString, etc.
- [x] Refactor value type optionals to use `*T` instead of `runtime.OptionalT` (per updated spec)

### Optional Types - Phases B & C
- [x] Phase B: `if x` checks - value types check `.Valid`, reference types check `!= nil`
- [x] Phase C: Assignment with check pattern `if (n = s.to_i?)` → `if n, ok := ...; ok`
- [x] `nil` literal for optional assignments
- [x] Optional methods `to_i?`/`to_f?` map to runtime functions

## Phase 12: Spec Compliance & Hardening ✅
- [x] Enforce single-entry rule: only one file with top-level statements per package (spec 2.1)
- [x] Implement `break` and `next` in functional blocks (spec 5.4)
  - [x] Update `runtime` iterative methods (`Each`, `Times`, etc.) to use `bool` return for control
  - [x] Update `runtime` transformation methods (`Map`, `Select`, etc.) to use `(T, bool)` return for control
  - [x] Update codegen to track "in block" state and emit `return true/false` instead of `break/continue`
  - [x] Add tests for `break`/`next` in various block types
- [x] Implement Statement Modifiers (`if`/`unless`) (spec 5.5)
  - [x] Add `unless` keyword support (lexer/parser)
  - [x] Update AST to include optional `Condition` and `IsUnless` flag on statements
  - [x] Update Parser to detect trailing `if` and `unless`
  - [x] Update Codegen to wrap statements in `if` blocks (inverting condition for `unless`)
- [x] Implement `unless` statement (spec 5.2)
  - [x] Reuse `IfStmt` with `IsUnless` flag
  - [x] Update Parser to handle `unless ... else ... end`
  - [x] Update Codegen to emit `if !cond`

## Phase 13: Symbols & Case Expressions
- [x] Implement Symbols (spec 4.1.1) ✅
  - [x] Add `SYMBOL` token type
  - [x] Update Lexer to recognize `:ident` as symbol (distinguishes from `: ` type annotation)
  - [x] Add `SymbolLit` AST node
  - [x] Update Parser to handle symbol literals in expressions
  - [x] Update Parser to support symbols in parenthesis-less calls (`puts :hello`)
  - [x] Update Codegen to emit string literals (`:foo` → `"foo"`)
  - [x] Update Codegen inferType to return "String" for symbols
  - [x] Add comprehensive tests (lexer, parser, codegen)
  - [x] Add examples/symbols.rg demonstration
- [x] Implement Case Expressions (spec 5.2) ✅
  - [x] Add `CASE`, `WHEN` tokens
  - [x] Add `CaseStmt` AST node with `Subject`, `WhenClauses`, `Else`
  - [x] Update Parser to handle `case ... when ... else ... end`
  - [x] Update Codegen to emit Go `switch` statements
  - [x] Support multiple values per when clause (`when 1, 2, 3`)
  - [x] Support case without subject (`case when x > 10`)
  - [x] Add comprehensive tests and examples/case.rg

## Phase 14: Class System Overhaul (v0.2)
Refining the class system based on Crystal's model.
- [ ] `attr_accessor` / `property` macros with types
- [ ] Class-level constants
- [ ] Static methods (`def self.method`)





