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

---

## Spec Implementation Progress

Tracking implementation of `spec.md` sections.

### Section 4: Lexical Structure ✅ COMPLETE
- [x] 4.1 Comments
- [x] 4.2 Identifiers (snake_case, predicates, setters, instance vars)
- [x] 4.3 Lambdas (`-> (x) { }` and `-> (x) do...end`)
- [x] 4.4 Strings (double quotes, interpolation, single quotes)
- [x] 4.5 Heredocs (`<<DELIM`, `<<-`, `<<~`, `<<'DELIM'`)
- [x] 4.6 Symbols (`:foo`, symbol keys in maps)
- [x] 4.7 Regular Expressions (`/pattern/`, `=~`, `!~`)

### Section 5: Types
- [x] 5.1 Primitive Types (Int, String, Bool, Float, etc.)
- [x] 5.2 Composite Types (Array, Map, Chan, Task)
- [x] 5.3 Type Aliases (`type UserID = Int64`)
- [ ] 5.4 Tuples (`(T, U)` types, destructuring)
- [x] 5.5 Array Literals
- [x] 5.6 Indexing (negative indices)
- [x] 5.7 Range Slicing (`arr[1..3]`, `arr[1..-1]`)
- [x] 5.8 Map Literals
- [ ] 5.9 Set Literals (`Set{1, 2, 3}`)
- [x] 5.10 Range Type
- [x] 5.11 Channel Type
- [x] 5.12 Task Type

### Section 6: Generics
- [ ] 6.1 Generic Functions (`def identity<T>(x : T) -> T`)
- [ ] 6.2 Generic Classes (`class Box<T>`)
- [ ] 6.3 Generic Interfaces
- [ ] 6.4 Type Constraints (`T : Comparable`)
- [ ] 6.5 Type Inference for Generics

### Section 7: Enums
- [ ] 7.1 Basic Enums
- [ ] 7.2 Enums with Explicit Values
- [ ] 7.3 Enum Methods (`to_s`, `values`, `from_string`)

### Section 8: Optionals ✅ COMPLETE
- [x] 8.1 Optional Operators (`??`, `&.`)
- [x] 8.2 Optional Methods (`ok?`, `unwrap`, `map`)
- [x] 8.3 Unwrapping Patterns (`if let`)
- [x] 8.4 The `nil` Keyword

### Section 9: Variables and Control Flow
- [x] 9.1 Variables (assignment, compound, `||=`)
- [x] 9.2 Constants
- [ ] 9.3 Destructuring (`a, b = pair`, `{name:} = map`)
- [x] 9.4 Conditionals (if/unless/elsif)
- [x] 9.5 Case Expressions
- [x] 9.6 Case Type (`case_type`)
- [x] 9.7 Loops (for/while/until/loop)
- [x] 9.8 Statement Modifiers
- [x] 9.9 Loop Modifiers

### Section 10: Functions
- [x] 10.1 Basic Functions
- [x] 10.2 Multiple Return Values
- [ ] 10.3 Default Parameters (`def foo(x = 10)`)
- [ ] 10.4 Named Parameters (`foo(timeout: 60)`)
- [ ] 10.5 Variadic Functions (`*args`)
- [x] 10.6 Lambdas
- [x] 10.7 Closures
- [x] 10.8 Symbol-to-Proc (`&:method`)
- [x] 10.9 Function Types
- [x] 10.10 Calling Convention

### Section 11: Classes ✅ COMPLETE
- [x] 11.1-11.7 All class features implemented

### Section 12: Structs
- [ ] 12.1 Struct Definition
- [ ] 12.2 Struct Features (auto-constructor, equality)
- [ ] 12.3 Struct Methods
- [ ] 12.4 Struct Immutability
- [ ] 12.5 Struct vs Class

### Section 13-14: Interfaces & Modules ✅ COMPLETE
- [x] Interfaces, structural typing, modules, mixins

### Section 15: Visibility ✅ COMPLETE
- [x] `pub`, private, naming transformations

### Section 16: Errors ✅ COMPLETE
- [x] Error signatures, `!` operator, `rescue`, custom errors

### Section 17: Concurrency ✅ COMPLETE
- [x] Goroutines, channels, select, spawn/await, concurrently

### Section 18: Go Interop ✅ COMPLETE
- [x] Imports, name mapping, defer

### Section 19: Runtime Package
- [x] Array methods (each, map, select, etc.)
- [x] Map methods
- [x] String methods
- [x] Integer methods (times, upto, downto)
- [x] Float methods
- [x] Global functions (puts, print, gets, etc.)

### Section 20-25: Testing & Appendices
- [ ] 20: Testing framework (describe, assertions)
- [x] Diagnostics, error messages

---

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

Current spec tests (58 total):
- `tests/spec/blocks/` - 9 tests (each, map_select, reduce, block_arithmetic, method_chaining_newlines, find_any_all_none, times_upto_downto, symbol_to_proc, multiple_params)
- `tests/spec/classes/` - 9 tests (basic, inheritance, inherited_getter, multilevel_inheritance, accessors, method_chaining, visibility, class_methods, super_calls)
- `tests/spec/concurrency/` - 6 tests (channels, goroutines, spawn_await, spawn_closure, concurrently, waitgroup)
- `tests/spec/control_flow/` - 7 tests (if_else, case_when, case_type, while_until, statement_modifiers, loop_modifiers, break_next)
- `tests/spec/errors/` - 3 tests (known limitations + runtime_panic)
- `tests/spec/error_handling/` - 2 tests (rescue, error_utilities)
- `tests/spec/functions/` - 1 test (basic)
- `tests/spec/go_interop/` - 1 test (strings)
- `tests/spec/interfaces/` - 2 tests (basic, any_indexing)
- `tests/spec/literals/` - 11 tests (arrays, integers, strings, ranges, range_include, empty_typed_array, map_symbol_shorthand, floats, heredocs, symbols, word_arrays)
- `tests/spec/modules/` - 3 tests (basic, multiple_includes, method_override)
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
- [x] Module method resolution (overriding)

### Concurrency (expand `tests/spec/concurrency/`)
- [x] Channel operations (`Chan[T].new`, send, receive)
- [x] Goroutines (`go do ... end`)
- [x] `spawn`/`await` blocks
- [x] `concurrently` blocks with scoped spawn
- [x] WaitGroup usage (via sync.WaitGroup)

### Error Handling (Go-style, not exceptions)
- [x] Inline rescue (`value = expr rescue default`)
- [x] Block rescue with error binding (`rescue => err do`)
- [x] Bang operator (`!`) for error propagation
- [x] `error_is?` and `error_as` utilities

---

## Phase 3: Compiler Architecture ✅ COMPLETE

### Separate Concerns in Codegen

**Goal:** Codegen should be a pure emitter that queries TypeInfo for all type/symbol information.

**Completed work:**
- [x] Tests run semantic analysis before codegen (via `compile()` helper)
- [x] Simplified `inferTypeFromExpr()` to rely on semantic type info
- [x] Extended TypeInfo interface with 20+ query methods
- [x] Removed fallback maps from codegen - helpers now panic if TypeInfo is nil
- [x] Removed pre-pass loops that populated fallback maps
- [x] Removed unused Generator fields: pubClasses, classes, classAccessorFields, interfaces, interfaceMethods, classConstructorParams, noArgFunctions, currentClassModuleMethods
- [x] Module accessor inheritance (propagate module accessors to including classes)
- [x] Module method origin tracking (`IsModuleMethod()` in TypeInfo)
- [x] Replaced `g.vars` type lookups with `TypeInfo.GetRugbyType()` where AST nodes available

**Acceptable emission-time state (must remain in codegen):**

| Field | Purpose | Why It's Emission State |
|-------|---------|------------------------|
| `g.vars` | Scoped variable tracking | Manages lexical scope during emission; handles save/restore for nested blocks |
| `g.goInteropVars` | Go interop type tracking | Control-flow sensitive - reassignment can remove tracking |
| `g.classMethods` | Generated Go function names | Maps Rugby method names to generated Go function names |
| `g.modules` | Module AST access | Needed for include processing during class emission |
| `g.accessorFields` | Current class field tracking | Tracks fields that need underscore prefix |
| `g.imports` | Import alias tracking | Tracks which Go packages are imported for interop detection |

**Test infrastructure:**
Consolidated test helpers into `codegen/helpers_test.go`:
- `compile()` - strict mode, fails on parser/semantic errors
- `compileRelaxed()` - ignores semantic errors (for testing codegen patterns)
- `compileExpectError()` - returns Generate() error
- `compileWithErrors()` - returns collected errors from gen.Errors()
- `compileWithLineDirectives()` - with source file for line directive tests
- `assertContains()`/`assertNotContains()` - output assertions

### Improve Semantic Analysis

- [x] Field inheritance propagation (getters/setters)
- [x] Track variable usage for unused variable detection
- [x] Resolve selector kinds (field/method/getter)
- [x] Module accessor inheritance (propagate module accessors to including classes)
- [x] Track module method origin (which methods come from included modules)
- [x] Circular inheritance detection with clear error messages
- [x] Interface validation considers inherited methods
- [x] Line fields added to TernaryExpr, BangExpr, IndexExpr, NilCoalesceExpr for better error positions
- [x] SpawnExpr infers Task element type from block return type
- [x] Array.map infers return type as Array[BlockReturnType]
- [x] Improved "Did you mean?" suggestions using Levenshtein distance
- [x] Reduce accumulator type inferred from initial value
- [x] Block return type inferred from control flow branches (if/else, return)
- [x] Inherited field types resolved via `getClassField()` helper

### Error Recovery

- [x] Parser: Continue after syntax errors to report multiple issues
- [x] Semantic: Collect all type errors before failing
- [x] Codegen: Collect errors via `addError()` instead of placeholder comments

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
- Features implemented: class methods, super calls, symbol-to-proc, spawn closure capture, module method overriding
- Semantic analysis: circular inheritance detection, inherited method interface validation, improved error positions, better type inference for spawn/map blocks, Levenshtein-based typo suggestions, reduce accumulator type inference, block return type inference from control flow, inherited field type resolution
- Error recovery: parser, semantic analyzer, and codegen all collect and report multiple errors
- Phase 2 complete: All spec test categories covered
- Phase 3 complete: TypeInfo migration, error collection
- Spec tests: 58 passing
- All `make check` passes
