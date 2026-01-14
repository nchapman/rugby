# Rugby Compiler TODO

Roadmap to a bulletproof compiler.

## Development Approach

### Spec-Driven Development (for new features)

> 📖 **Full guide:** See [`docs/guide.md`](docs/guide.md) for detailed patterns and examples.

When implementing new language features, follow this workflow:

1. **Spec test first** → `tests/spec/errors/<feature>.rg` with `#@ compile-fail`
2. **AST** → Add/modify nodes in `ast/ast.go`
3. **Parser** → Recognize syntax in `parser/`, build AST
4. **Semantic** → Type inference + validation in `semantic/` (this is where the magic happens)
5. **CodeGen** → Emit Go code in `codegen/`, querying `typeInfo` (no inference!)
6. **Flip test** → Move to feature directory, change to `#@ run-pass`

**The golden rule:** Semantic analysis owns all type inference. CodeGen is a pure emitter.

```
Parser → AST → Semantic (types) → CodeGen (emit)
                   ↓                  ↑
              nodeTypes ──────► typeInfo queries
```

#### Feature Priority

| Feature | Value | Complexity | Notes |
|---------|-------|------------|-------|
| ~~Super calls with args~~ | ~~High~~ | ~~Medium~~ | ✅ Done |
| ~~Symbol-to-proc (`&:method`)~~ | ~~Medium~~ | ~~Low-Medium~~ | ✅ Done |
| ~~Spawn closure capture~~ | ~~Medium~~ | ~~Medium~~ | ✅ Works |

### Documenting Limitations (for workarounds)

When writing spec tests and discovering limitations:

1. **Document the limitation** in the "Known Limitations" section below
2. **Write the test using a valid workaround** (e.g., explicit `return` instead of implicit)
3. **Add a `# TODO:` comment** in the test file describing the ideal syntax
4. **Later**: Fix the limitation, then add a new test variation with the ideal syntax

This approach keeps momentum while building a complete picture of what needs fixing.
To find all limitation workarounds: `grep -r "# TODO:" tests/spec/`

---

## Phase 1: Fix Active Bugs ✅ COMPLETE

All major bugs have been fixed. The remaining items are documented limitations.

### Fixed Bugs

- [x] **BUG-038: Method chaining with newlines** - Works correctly
- [x] **BUG-046: Map literal in method body** - Works correctly
- [x] **BUG-036: Empty typed array literals** - Works correctly
- [x] **BUG-048: Interface indexing** - Fixed with reflection in `runtime.GetKey()`
- [x] **Block variables typed as any** - Fixed with native for-range loops
- [x] **Range.include? returns Range** - Fixed parser handling of parenthesized expressions
- [x] **Inherited getters return pointers** - Fixed with `propagateInheritedFields()` in semantic analyzer
- [x] **Array.first/last return pointers** - Fixed return type tracking in semantic analyzer

### Known Limitations

- [ ] **Multi-line if expression** (`tests/spec/errors/if_expression_multiline.rg`)
  - `result = if x > 5\n  "yes"\nelse\n  "no"\nend` requires inline form
  - Workaround: Use single-line `result = if x > 5 then "yes" else "no" end` or ternary `x > 5 ? "yes" : "no"`

- [ ] **Inline type annotations** (`tests/spec/errors/type_annotation_inline.rg`)
  - `[] : Array[String]` inline annotation not yet supported
  - Workaround: Use variable annotation `arr : Array[String] = []`

- [ ] **Case/when implicit returns**
  - `case x when 1 then "one" end` doesn't implicitly return the value
  - In Ruby, `case/when` is an expression that returns the last value of the matched branch
  - Workaround: Use explicit `return` statements in each branch

- [ ] **Compound assignment in loop modifier expressions**
  - `puts counter += 1 while counter < 3` fails to parse
  - The parser doesn't support compound assignment as part of an expression
  - Workaround: Use regular while loop with compound assignment in body

- [ ] **Range slice returns any**
  - `arr[1..-1]` returns `any` type, can't reassign to typed variable
  - Workaround: Use explicit loop or different approach for slicing

- [x] **Super calls with arguments** ✅
  - `super(args)` now passes arguments to parent constructor
  - `super(args)` in method body calls parent method correctly
  - Inherited methods can be called without parentheses

- [x] **Spawn blocks capture outer variables** ✅
  - `spawn { outer_var + 1 }` works correctly
  - Variables from enclosing scope are captured in spawn blocks

- [ ] **Array mutation in closures doesn't persist**
  - `arr << value` inside blocks doesn't modify the outer array
  - The `<<` operator generates `runtime.ShiftLeft` instead of append
  - Workaround: Use return values instead of mutation

---

## Phase 2: Expand Spec Test Coverage

Goal: Every language feature has spec tests covering all syntactic variations.

Current spec tests (56 total):
- `tests/spec/blocks/` - 9 tests (each, map_select, reduce, block_arithmetic, method_chaining_newlines, find_any_all_none, times_upto_downto, symbol_to_proc, multiple_params)
- `tests/spec/classes/` - 9 tests (basic, inheritance, inherited_getter, multilevel_inheritance, accessors, method_chaining, visibility, class_methods, super_calls)
- `tests/spec/concurrency/` - 4 tests (channels, goroutines, spawn_await, spawn_closure)
- `tests/spec/control_flow/` - 7 tests (if_else, case_when, case_type, while_until, statement_modifiers, loop_modifiers, break_next)
- `tests/spec/errors/` - 2 tests (known limitations + runtime_panic)
- `tests/spec/error_handling/` - 2 tests (rescue, error_utilities)
- `tests/spec/functions/` - 1 test (basic)
- `tests/spec/go_interop/` - 1 test (strings)
- `tests/spec/interfaces/` - 2 tests (basic, any_indexing)
- `tests/spec/literals/` - 11 tests (arrays, integers, strings, ranges, range_include, empty_typed_array, map_symbol_shorthand, floats, heredocs, symbols, word_arrays)
- `tests/spec/modules/` - 2 tests (basic, multiple_includes)
- `tests/spec/optionals/` - 3 tests (basic, if_let, nil_coalescing)
- `tests/spec/stdlib/` - 1 test (regex)

### Literals (expand `tests/spec/literals/`)
- [x] Floats (basic operations, predicates, rounding)
- [x] Heredocs (`<<DELIM`)
- [x] Word arrays (`%w[a b c]`, `%w(...)`)
- [x] Symbol literals (`:foo`, symbol in maps)
- [x] Regex via `rugby/regex` module (no literal syntax)

### Control Flow (expand `tests/spec/control_flow/`)
- [x] Case/when statements
- [x] Statement modifiers (`puts x if condition`)
- [x] While/until loops
- [x] Loop control (`break`, `next`)
- [ ] Begin/rescue/ensure

### Classes (expand `tests/spec/classes/`)
- [x] Property declarations (`property`, `getter`, `setter`)
- [x] Visibility (`pub` class and methods)
- [x] Method chaining with `self` return
- [x] Class methods (`def self.method`)
- [x] Super calls in methods

### Blocks (expand `tests/spec/blocks/`)
- [x] Iterator methods (`find`, `any?`, `all?`, `none?`)
- [x] `times`, `upto`, `downto`
- [x] Block with multiple parameters
- [x] Symbol-to-proc (`&:method`)

### Modules
- [x] Module definition and include
- [x] Multiple module includes
- [ ] Module method resolution (overriding)

### Concurrency (expand `tests/spec/concurrency/`)
- [x] Channel operations (`Chan[T].new`, send, receive)
- [x] Goroutines (`go do ... end`)
- [x] `spawn`/`await` blocks
- [ ] `concurrently` blocks with scoped spawn
- [ ] WaitGroup usage

### Error Handling (Go-style, not exceptions)
- [x] Inline rescue (`value = expr rescue default`)
- [x] Block rescue with error binding (`rescue => err do`)
- [x] Bang operator (`!`) for error propagation
- [x] `error_is?` and `error_as` utilities

---

## Phase 3: Compiler Architecture

### Separate Concerns in Codegen

Currently `codegen/codegen.go` Generator struct has 35+ fields mixing:
- Symbol tables (`vars`, `classes`, `interfaces`)
- Type information (`classFields`, `accessorFields`)
- Go interop tracking (`goInteropVars`, `imports`)
- Emission state (`buf`, `indent`, `currentClass`)

**Goal:** Codegen should be a pure emitter that receives fully-analyzed AST.

- [ ] Move symbol tables to semantic analyzer
- [ ] Expand `TypeInfo` interface to provide all resolution info
- [ ] Remove type inference from codegen
- [ ] Codegen only transforms AST nodes to Go syntax

### Improve Semantic Analysis

- [x] Field inheritance propagation (getters/setters)
- [x] Track variable usage for unused variable detection
- [x] Resolve selector kinds (field/method/getter)
- [ ] Complete type inference for all expressions
- [ ] Validate interface satisfaction at analysis time
- [ ] Better error messages with source context

### Error Recovery

- [ ] Parser: Continue after syntax errors to report multiple issues
- [ ] Semantic: Collect all type errors before failing
- [ ] Codegen: Generate placeholder code for unresolved symbols

---

## Phase 4: Language Completeness

### Missing Ruby Features (evaluate priority)
- [ ] Default parameter values (`def foo(x = 10)`)
- [ ] Splat operators (`*args`, `**kwargs`)
- [ ] Destructuring assignment (`a, b = [1, 2]`)
- [ ] Pattern matching (`case x in ...`)
- [ ] String formatting (`%` operator, `sprintf`)

### Go Interop Improvements
- [ ] Automatic interface implementation detection
- [ ] Better handling of Go error returns
- [ ] Support for Go generics
- [ ] Embedding Go structs

### Tooling
- [ ] Language server (LSP) for IDE support
- [ ] Formatter (`rugby fmt`)
- [ ] Documentation generator

---

## Phase 5: Quality & Performance

### Testing
- [ ] Fuzz testing for parser
- [ ] Property-based tests for type system
- [ ] Benchmark suite for compilation speed
- [ ] Memory profiling for large files

### Error Messages
- [ ] Source snippets in all errors
- [ ] Suggestions for common mistakes
- [ ] "Did you mean?" for typos

### Optimization
- [ ] Parallel compilation of independent files
- [ ] Incremental compilation (only recompile changed)
- [ ] Generated code optimization passes

---

## Tracking Progress

Run spec tests to see current status:
```bash
make test    # Run all tests
go test ./tests/... -run TestSpecs -v  # Run only spec tests
```

When fixing a bug:
1. Find its test in `tests/spec/errors/`
2. Fix the compiler
3. Move test to feature directory, change `#@ compile-fail` to `#@ run-pass`
4. Add `#@ check-output` with expected output
5. Run `make check` to verify

Current status:
- Original bugs: 8 fixed, 2 documented as limitations (multi-line if, inline type annotations)
- Additional limitations discovered: 4 (case/when implicit returns, compound assignment in loop modifiers, range slice returns any, array mutation in closures)
- Features implemented: class methods, super calls, symbol-to-proc, spawn closure capture
- Spec tests: 56 passing
- All `make check` passes
