# Rugby MVP TODO

Features in implementation order, building incrementally.

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

**Note on snake_case transformation (see spec.md Section 5):**
Currently, snake_case → PascalCase only applies to selector calls (Go interop).
Local function names are passed through as-is. Full behavior once `pub` is implemented:
- Selector calls (`io.read_all`) → PascalCase (`io.ReadAll`) - Go exports are always uppercase
- `pub def parse_json` → `func ParseJSON()` - exported
- `def parse_json` → `func parseJSON()` - internal (camelCase)
- Acronym handling (5.3): `id` → `ID`, `http` → `HTTP`, etc.
- Name collision detection (5.4): compile-time errors for collisions
- Reserved word escaping (5.5): suffix `_` for internal, error for `pub`
This requires `pub` keyword support (Phase 8) to be fully correct.

## Phase 5: Collections
- [ ] Array literals (`[1, 2, 3]`)
- [ ] Array indexing (`arr[0]`)
- [ ] Map literals (`{"a" => 1}`)
- [ ] Map access (`m["key"]`)
- [ ] `each` blocks (`arr.each do |x| ... end`)
- [ ] `each_with_index`
- [ ] `map` blocks

## Phase 6: Classes
- [ ] Class definition (`class User ... end`)
- [ ] Instance variables (`@name`)
- [ ] Constructor (`def initialize`)
- [ ] `Class.new(...)` syntax
- [ ] Methods with receivers
- [ ] Pointer receiver methods (`def mutate!`)
- [ ] Embedding via `<` (`class Service < Logger`)

## Phase 7: Types & Optionals
- [ ] Type annotations (`x : Int = 5`)
- [ ] Parameter type annotations (`def foo(x : String)`)
- [ ] Optional type (`T?`)
- [ ] `nil` literal
- [ ] Optional unwrapping in conditionals (`if (x = maybe_val?)`)

## Phase 8: Interfaces & Visibility
- [ ] Interface definitions (`interface Speaker`)
- [ ] Structural conformance checking
- [ ] `pub` keyword for exports

## Phase 9: Strings & Polish
- [ ] String interpolation (`"hi #{name}"`)
- [ ] Comments in more positions
- [ ] Better error messages with source locations
- [ ] Multi-file compilation
