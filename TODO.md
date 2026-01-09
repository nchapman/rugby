# Rugby MVP TODO

Features in implementation order, building incrementally.

**Recent spec updates (see spec.md):**
- Return statements (6.3): explicit + implicit returns
- Block semantics (5.4): iteration-only for MVP, not closures
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

## Phase 5: Collections
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
  - [x] Update codegen: recognize patterns (emit `for range`) vs IIFE for map
  - [x] Proper variable scope tracking in blocks
  - [x] Duplicate block parameter detection
  - [x] Nested blocks support
  - [ ] Create `rugby` runtime package for stdlib functions
  - [ ] Add more array methods: `select`, `reject`, `reduce`, `find`, etc.

## Phase 6: Classes
- [ ] Class definition (`class User ... end`)
- [ ] Instance variables (`@name`) - types inferred from `initialize`
- [ ] Constructor (`def initialize`) - generates `New*` functions
- [ ] `Class.new(...)` syntax
- [ ] Methods with receivers (value receiver default)
- [ ] Pointer receiver methods (`def mutate!`)
- [ ] Embedding via `<` (`class Service < Logger`) - single parent only for MVP

## Phase 7: Types & Optionals
- [ ] Type annotations (`x : Int = 5`)
- [ ] Parameter type annotations (`def foo(x : String)`)
- [ ] Optional type (`T?`) - value types only; compiles to `(T, bool)`
- [ ] `nil` literal - for reference types only
- [ ] Optional unwrapping in conditionals (`if (x = maybe_val?)`)

## Phase 8: Interfaces & Visibility
- [ ] Interface definitions (`interface Speaker`) - methods implicitly exported
- [ ] Structural conformance checking
- [ ] `pub` keyword for exports - enables proper name casing

## Phase 9: Strings & Polish
- [ ] String interpolation (`"hi #{name}"`) - any expression, uses `fmt.Sprintf`
- [ ] Comments in more positions (currently only full-line comments)
- [ ] Better error messages with source locations
- [ ] Multi-file compilation
