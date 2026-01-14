# Rugby Compiler TODO

Roadmap to a bulletproof compiler.

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

---

## Phase 2: Expand Spec Test Coverage

Goal: Every language feature has spec tests covering all syntactic variations.

Current spec tests (27 total):
- `tests/spec/blocks/` - 5 tests (each, map_select, reduce, block_arithmetic, method_chaining)
- `tests/spec/classes/` - 4 tests (basic, inheritance, inherited_getter, multilevel_inheritance)
- `tests/spec/control_flow/` - 1 test (if_else)
- `tests/spec/errors/` - 3 tests (known limitations + runtime_panic)
- `tests/spec/functions/` - 1 test (basic)
- `tests/spec/go_interop/` - 1 test (strings)
- `tests/spec/interfaces/` - 2 tests (basic, any_indexing)
- `tests/spec/literals/` - 6 tests (arrays, integers, strings, ranges, range_include, empty_typed_array, map_symbol_shorthand)
- `tests/spec/optionals/` - 3 tests (basic, if_let, nil_coalescing)

### Literals (expand `tests/spec/literals/`)
- [ ] Floats (scientific notation, edge cases)
- [ ] Heredocs (`<<EOF`, `<<-EOF`, `<<~EOF`)
- [ ] Word arrays (`%w[a b c]`)
- [ ] Symbol literals (`:foo`, `:"foo bar"`)
- [ ] Regex literals (`/pattern/`)

### Control Flow (expand `tests/spec/control_flow/`)
- [ ] Case/when statements
- [ ] Statement modifiers (`puts x if condition`)
- [ ] While/until loops
- [ ] Loop control (`break`, `next`, `redo`)
- [ ] Begin/rescue/ensure

### Classes (expand `tests/spec/classes/`)
- [ ] Property declarations (`property`, `getter`, `setter`)
- [ ] Class methods (`def self.method`)
- [ ] Visibility (`pub`, `private`)
- [ ] Method chaining with `self` return
- [ ] Super calls in methods

### Blocks (expand `tests/spec/blocks/`)
- [ ] All iterator methods (`find`, `any?`, `all?`, `none?`, `count`)
- [ ] `times`, `upto`, `downto`
- [ ] Block with multiple parameters
- [ ] Symbol-to-proc (`&:method`)

### Modules
- [ ] Module definition and include
- [ ] Module method resolution
- [ ] Multiple module includes

### Concurrency (create `tests/spec/concurrency/`)
- [ ] Channel operations (`Chan[T].new`, send, receive)
- [ ] `spawn` blocks
- [ ] `concurrently` blocks with `await`
- [ ] WaitGroup usage

### Error Handling
- [ ] Begin/rescue/ensure
- [ ] Typed rescue (`rescue TypeError => e`)
- [ ] Raise/throw
- [ ] Custom error classes

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
- Bugs fixed: 8/10 (2 are documented limitations)
- Spec tests: 27 passing
- All `make check` passes
