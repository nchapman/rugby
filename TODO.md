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
- [ ] Explicit return types (`def add(a, b) -> Int`)
- [ ] Multiple return values (`-> (Int, Bool)`)
- [x] Function calls with arguments (`add(1, 2)`)

## Phase 4: Go Interop
- [ ] Method calls with dot syntax (`http.Get(url)`)
- [ ] Import aliases (`import encoding/json as json`)
- [ ] snake_case to CamelCase mapping (`read_all` -> `ReadAll`)
- [ ] Defer (`defer f.Close`)

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
