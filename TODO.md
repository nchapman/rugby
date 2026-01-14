# Rugby Compiler TODO

Roadmap to a bulletproof compiler.

## Phase 1: Fix Active Bugs

These are documented as failing spec tests in `tests/spec/errors/`. When fixed, move the test to the appropriate feature directory and change to `#@ run-pass`.

### Parser Bugs

- [ ] **BUG-038: Method chaining with newlines** (`method_chaining_newlines.rg`)
  - `[1,2,3]\n  .map { }` fails - parser sees `.map` as statement start
  - Fix: Allow `.` at line start to continue previous expression

- [ ] **BUG-046: Map literal in method body** (`map_literal_method.rg`)
  - `{ id: @id }` symbol shorthand fails in method context
  - Fix: Disambiguate map literal vs block in method body

- [ ] **If expression spanning multiple lines** (`if_expression_multiline.rg`)
  - `result = if x > 5\n  "yes"\nelse\n  "no"\nend` fails
  - Fix: Allow multi-line if-as-expression in assignment context

### Type System Bugs

- [ ] **BUG-036: Empty typed array literals** (`empty_typed_array.rg`, `type_annotation_inline.rg`)
  - `empty : Array[Int] = []` gives type mismatch
  - `[] : Array[String]` inline annotation unrecognized
  - Fix: Propagate type context to empty array literals

- [ ] **BUG-048: Interface indexing** (`interface_indexing.rg`)
  - `post["title"]` fails when `post` is type `any`
  - Fix: Allow indexing on `any` type, generate type assertion

- [ ] **Block variables typed as any** (`block_arithmetic.rg`)
  - `nums.each { |n| n * 10 }` fails - `n` is `any` not `Int`
  - Fix: Infer block parameter type from array element type

- [ ] **Range.include? returns Range** (`range_include_condition.rg`)
  - `if (1..10).include?(5)` fails - returns Range not Bool
  - Fix: Correct return type for range predicate methods

### Codegen Bugs

- [ ] **Inherited getters return pointers** (`inherited_getter_pointer.rg`)
  - `dog.name` returns pointer address when `name` is inherited getter
  - Fix: Generate proper accessor call for inherited getters

- [ ] **Array.first/last return pointers** (`array_first_last_pointer.rg`)
  - `nums.first` returns `*int` instead of `int`
  - Fix: Dereference in runtime or codegen

---

## Phase 2: Expand Spec Test Coverage

Goal: Every language feature has spec tests covering all syntactic variations.

### Literals (expand `tests/spec/literals/`)
- [ ] Floats (scientific notation, edge cases)
- [ ] Heredocs (`<<EOF`, `<<-EOF`, `<<~EOF`)
- [ ] Word arrays (`%w[a b c]`)
- [ ] Symbol literals (`:foo`, `:"foo bar"`)
- [ ] Regex literals (`/pattern/`)
- [ ] Map literals (all forms: `{}`, `{ a: 1 }`, `{ "a" => 1 }`)

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
- [ ] Super calls

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

- [ ] Complete type inference for all expressions
- [ ] Track variable usage for unused variable detection
- [ ] Resolve all method calls (distinguish field/method/getter)
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
make test-spec
```

When fixing a bug:
1. Find its test in `tests/spec/errors/`
2. Fix the compiler
3. Move test to feature directory, change `#@ compile-fail` to `#@ run-pass`
4. Add `#@ check-output` with expected output
5. Run `make test-spec` to verify

Bug count: `ls tests/spec/errors/*.rg | wc -l`
