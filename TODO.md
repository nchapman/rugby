# Rugby Compiler TODO

## How to Use This File

This file tracks the implementation of `spec.md`. The spec tests in `tests/spec/` are written to match the spec exactly.

### Workflow

1. **Pick a section** - Start from the top, work down. Earlier sections are dependencies for later ones.
2. **Run the tests** - `make test-spec` to see what's failing.
3. **Fix the compiler** - Make the tests pass.
4. **Get a review** - Run the code-reviewer agent before committing.
5. **Commit** - Small, focused commits as you complete each subsection.

### Source of Truth

**`spec.md` is the source of truth.** If a test fails, the compiler needs to change, not the test. If you're unsure about behavior, check `spec.md`.

### Running Tests

```bash
make test-spec                    # Run all spec tests
go test ./tests/... -run TestSpecs/types -v   # Run one category
go test ./tests/... -run "TestSpecs/types/composite_array" -v  # Run one test
```

### Test Directives

| Directive | Meaning |
|-----------|---------|
| `#@ run-pass` | Must compile and run |
| `#@ compile-fail` | Must fail to compile |
| `#@ skip: reason` | Skip (unimplemented feature) |
| `#@ check-output` | Verify stdout matches `#@ expect:` |

---

## Phase 1: Type Syntax (Foundation)

The new spec uses angle bracket syntax for generic types. This must be implemented first as everything else depends on it.

### 1.1 Generic Type Syntax in Parser

Update the parser to recognize angle bracket type parameters:

- [ ] `Array<T>` - was `Array[T]`
- [ ] `Map<K, V>` - was `Map[K, V]`
- [ ] `Set<T>` - was `Set[T]`
- [ ] `Chan<T>` - was `Chan[T]`
- [ ] `Task<T>` - was `Task[T]`
- [ ] Nested types: `Array<Map<String, Int>>`
- [ ] Optional generic: `Array<Int>?`

**Tests:** `tests/spec/types/composite_*.rg`

### 1.2 Type Annotations

- [ ] Variable type annotations: `x : Array<Int> = []`
- [ ] Function parameter types: `def foo(arr : Array<String>)`
- [ ] Return type annotations: `def foo -> Map<String, Int>`

**Tests:** `tests/spec/types/*.rg`

---

## Phase 2: Primitives and Literals

### 2.1 Primitive Types

- [x] Int, Int8, Int16, Int32, Int64
- [x] UInt, UInt8, UInt16, UInt32, UInt64
- [x] Float, Float32
- [x] Bool
- [x] String
- [ ] Bytes (verify `"str".bytes` works)
- [ ] Rune

**Tests:** `tests/spec/types/primitive_*.rg`

### 2.2 String Literals

- [x] Double-quoted strings with interpolation
- [ ] Single-quoted strings (no interpolation, only `\'` and `\\` escapes)
- [x] String interpolation `#{expr}`
- [x] Escape sequences

**Tests:** `tests/spec/literals/strings.rg`, `single_quote_strings.rg`, `string_interpolation.rg`

### 2.3 Heredocs

- [x] Basic heredoc `<<DELIM`
- [x] Indented delimiter `<<-DELIM`
- [ ] Squiggly heredoc `<<~DELIM` (strips leading whitespace)
- [ ] Literal heredoc `<<'DELIM'` (no interpolation)

**Tests:** `tests/spec/literals/heredoc_*.rg`

### 2.4 Other Literals

- [x] Integer literals
- [x] Float literals
- [x] Array literals `[1, 2, 3]`
- [x] Map literals `{key: value}` and `{"key" => value}`
- [x] Set literals `Set{1, 2, 3}`
- [x] Range literals `1..5`, `1...5`
- [x] Symbol literals `:foo`
- [x] Word arrays `%w[a b c]`

**Tests:** `tests/spec/literals/*.rg`

---

## Phase 3: Variables and Control Flow

### 3.1 Variables

- [x] Basic assignment `x = expr`
- [x] Type-annotated assignment `x : Type = expr`
- [x] Compound operators `+=`, `-=`, `*=`, `/=`, `%=`
- [ ] `||=` for optionals only

**Tests:** `tests/spec/control_flow/variables.rg`, `or_assign.rg`

### 3.2 Constants

- [x] `const NAME = value`
- [x] Typed constants `const NAME : Type = value`
- [x] Computed constants `const SIZE = OTHER * 2`

**Tests:** `tests/spec/control_flow/constants.rg`

### 3.3 Conditionals

- [x] `if`/`elsif`/`else`/`end`
- [x] `unless`
- [x] Ternary `a ? b : c`
- [x] Conditions must be Bool (no implicit truthiness)

**Tests:** `tests/spec/control_flow/conditionals.rg`, `if_else.rg`

### 3.4 Case Expressions

- [x] `case`/`when`/`else`/`end`
- [x] Multiple values `when 1, 2, 3`
- [x] Range matching `when 1..10`
- [ ] `case_type` for type switching

**Tests:** `tests/spec/control_flow/case_*.rg`

### 3.5 Loops

- [x] `for item in collection`
- [x] `for i in range`
- [x] `for key, value in map`
- [x] `while condition`
- [x] `until condition`
- [x] `loop do ... end`
- [x] `break`, `next`

**Tests:** `tests/spec/control_flow/loops_*.rg`, `break_next.rg`

### 3.6 Statement Modifiers

- [x] `return x if condition`
- [x] `puts x unless condition`
- [ ] `expr while condition` (loop modifier)
- [ ] `expr until condition` (loop modifier)

**Tests:** `tests/spec/control_flow/statement_modifiers*.rg`, `loop_modifiers*.rg`

---

## Phase 4: Functions

### 4.1 Basic Functions

- [x] `def name(params) ... end`
- [x] Explicit return `return value`
- [x] Implicit return (last expression)
- [x] Type annotations on params and return

**Tests:** `tests/spec/functions/basic*.rg`

### 4.2 Multiple Returns (Tuples)

- [x] `def foo -> (T, U)`
- [x] `return a, b`
- [x] Destructuring `x, y = foo()`
- [x] Ignore with `_`

**Tests:** `tests/spec/functions/multiple_returns.rg`, `tests/spec/types/tuples*.rg`

### 4.3 Lambdas

- [x] Single-line `-> (x) { expr }`
- [x] Multi-line `-> (x) do ... end`
- [x] With return type `-> (x) -> T do ... end`
- [x] Calling with `.()` or `.call()`
- [x] Lambda `return` returns from lambda (not enclosing function)

**Tests:** `tests/spec/functions/lambda_*.rg`

### 4.4 Closures

- [x] Capture variables by reference
- [x] Mutations affect original variable

**Tests:** `tests/spec/functions/closures.rg`

### 4.5 Symbol-to-Proc

- [x] `&:method` syntax
- [x] Works with map, select, reject, etc.

**Tests:** `tests/spec/functions/symbol_to_proc*.rg`, `tests/spec/blocks/symbol_to_proc.rg`

### 4.6 Function Types

- [x] `type Handler = (Req) -> Resp`
- [x] Functions as parameters

**Tests:** `tests/spec/functions/function_types.rg`

---

## Phase 5: Classes

### 5.1 Class Basics

- [x] `class Name ... end`
- [x] `def initialize(@field : Type)`
- [x] Instance variables `@field`
- [x] Parameter promotion in initialize

**Tests:** `tests/spec/classes/basic.rg`, `instance_variables.rg`

### 5.2 Accessors

- [x] `getter name : Type`
- [x] `setter name : Type`
- [x] `property name : Type`
- [ ] `pub getter`, `pub setter`, `pub property`
- [x] Custom setters `def name=(value)`

**Tests:** `tests/spec/classes/accessors*.rg`, `custom_accessors.rg`

### 5.3 Class Methods

- [x] `def self.method_name`
- [x] Class variables `@@var`

**Tests:** `tests/spec/classes/class_methods*.rg`

### 5.4 Inheritance

- [x] `class Child < Parent`
- [x] `super` and `super(args)`
- [x] Method overriding

**Tests:** `tests/spec/classes/inheritance*.rg`, `super_calls.rg`

---

## Phase 6: Interfaces

### 6.1 Interface Definition

- [x] `interface Name ... end`
- [x] Method signatures
- [x] Interface composition `interface A < B, C`

**Tests:** `tests/spec/interfaces/interface_*.rg`

### 6.2 Structural Typing

- [x] Classes satisfy interfaces automatically
- [x] `implements` for compile-time verification

**Tests:** `tests/spec/interfaces/structural_typing.rg`

### 6.3 Type Checking

- [x] `obj.is_a?(Type)` returns Bool
- [x] `obj.as(Type)` returns Type?
- [ ] Works in `if let`

**Tests:** `tests/spec/interfaces/type_checking.rg`

### 6.4 Any Type

- [x] `Any` for heterogeneous collections
- [x] Must use `as` to recover concrete type

**Tests:** `tests/spec/interfaces/any_type.rg`, `any_indexing.rg`

---

## Phase 7: Modules

### 7.1 Module Definition

- [x] `module Name ... end`
- [x] `include ModuleName`
- [x] Module methods become instance methods

**Tests:** `tests/spec/modules/module_definition.rg`, `basic.rg`

### 7.2 Module State

- [ ] Module instance variables
- [ ] `||=` in modules

**Tests:** `tests/spec/modules/module_state.rg`

### 7.3 Module Namespacing

- [x] `Module::NestedClass`
- [x] `def self.method` for module-level functions

**Tests:** `tests/spec/modules/module_namespace.rg`

### 7.4 Conflict Resolution

- [x] Later includes override earlier
- [x] Class methods override included methods

**Tests:** `tests/spec/modules/module_conflicts.rg`, `method_override.rg`

---

## Phase 8: Optionals

### 8.1 Optional Type

- [x] `T?` type annotation
- [x] `nil` keyword

**Tests:** `tests/spec/optionals/basic.rg`, `nil_keyword.rg`

### 8.2 Nil Coalescing

- [x] `expr ?? default`
- [x] Checks presence, not truthiness (0 and "" are present)

**Tests:** `tests/spec/optionals/nil_coalescing*.rg`

### 8.3 Safe Navigation

- [x] `obj&.method`
- [x] Returns `R?`

**Tests:** `tests/spec/optionals/safe_navigation.rg`

### 8.4 If Let

- [x] `if let x = optional_expr`
- [x] Scoped binding (shadows outer variable)
- [x] With else clause

**Tests:** `tests/spec/optionals/if_let*.rg`

### 8.5 Optional Methods

- [ ] `ok?` / `present?`
- [ ] `nil?` / `absent?`
- [ ] `unwrap` (panics if nil)
- [ ] `unwrap_or(default)`
- [ ] `map`, `flat_map`, `each`, `filter`

**Tests:** `tests/spec/optionals/optional_methods.rg`

---

## Phase 9: Error Handling

### 9.1 Error Type

- [x] `Error` type (Go's `error`)
- [x] `(T, Error)` return pattern
- [x] `errors.new("message")`

**Tests:** `tests/spec/error_handling/error_signatures.rg`

### 9.2 Custom Errors

- [x] Class with `def error -> String`

**Tests:** `tests/spec/error_handling/custom_errors.rg`

### 9.3 Bang Operator

- [x] `expr!` propagates error to caller
- [x] Enclosing function must return Error
- [ ] In main/scripts, prints and exits

**Tests:** `tests/spec/error_handling/bang_operator.rg`

### 9.4 Rescue

- [x] Inline `expr rescue default`
- [x] Block `expr rescue do ... end`
- [x] With binding `expr rescue => err do ... end`

**Tests:** `tests/spec/error_handling/rescue_*.rg`

---

## Phase 10: Visibility

### 10.1 Pub Modifier

- [x] `pub def` exports function
- [x] `pub class` exports class
- [ ] `pub getter`, `pub setter`, `pub property`

**Tests:** `tests/spec/visibility/pub_export.rg`

### 10.2 Private Modifier

- [ ] `private def` restricts to class

**Tests:** `tests/spec/visibility/class_visibility.rg`

### 10.3 Name Transformations

- [x] snake_case → camelCase (internal)
- [x] snake_case → CamelCase (pub)
- [x] Acronym handling (userID, httpClient)

**Tests:** `tests/spec/visibility/naming_transform.rg`

---

## Phase 11: Concurrency

### 11.1 Goroutines

- [x] `go expr`
- [x] `go do ... end`

**Tests:** `tests/spec/concurrency/goroutines*.rg`

### 11.2 Channels

- [x] `Chan<T>.new(buffer_size)`
- [x] `ch << value` send
- [x] `ch.receive` blocking receive
- [ ] `ch.try_receive` non-blocking
- [x] `ch.close`
- [x] `for msg in ch` iterate

**Tests:** `tests/spec/concurrency/channels*.rg`

### 11.3 Select

- [x] `select`/`when`/`else`/`end`
- [x] Receive case `when val = ch.receive`
- [x] Send case `when ch << value`

**Tests:** `tests/spec/concurrency/select*.rg`

### 11.4 Spawn/Await

- [x] `spawn { expr }` returns `Task<T>`
- [x] `await task` blocks until complete
- [x] Closure capture in spawn

**Tests:** `tests/spec/concurrency/spawn_*.rg`

### 11.5 Concurrently

- [x] `concurrently -> (scope) do ... end`
- [x] `scope.spawn { }`
- [x] Cleanup on block exit

**Tests:** `tests/spec/concurrency/concurrently*.rg`

### 11.6 Sync Primitives

- [x] `sync.Mutex`
- [x] `sync.WaitGroup`
- [x] `sync.Once`

**Tests:** `tests/spec/concurrency/mutex*.rg`, `waitgroup.rg`, `once*.rg`

---

## Phase 12: Go Interop

### 12.1 Imports

- [x] `import "package"`
- [x] `import "package" as alias`

**Tests:** `tests/spec/go_interop/imports.rg`, `import_alias.rg`

### 12.2 Name Mapping

- [x] `snake_case` → `CamelCase` for Go calls
- [x] Automatic for function calls on imported packages

**Tests:** `tests/spec/go_interop/name_mapping.rg`

### 12.3 Defer

- [x] `defer expr`
- [x] LIFO execution order

**Tests:** `tests/spec/go_interop/defer*.rg`

---

## Phase 13: Runtime Methods

### 13.1 Array Methods

- [x] `each`, `each_with_index`
- [x] `map`, `select`, `reject`
- [x] `find`, `any?`, `all?`, `none?`
- [x] `reduce`, `sum`, `min`, `max`
- [x] `first`, `last`, `take`, `drop`
- [x] `length`, `size`, `empty?`
- [ ] `include?`, `contains?`
- [ ] `compact`, `uniq`, `flatten`
- [ ] `sorted`, `reversed`

**Tests:** `tests/spec/runtime/array_*.rg`, `tests/spec/blocks/*.rg`

### 13.2 Map Methods

- [x] `each`, `each_key`, `each_value`
- [x] `keys`, `values`
- [x] `length`, `size`, `empty?`
- [x] `has_key?`, `key?`
- [x] `fetch`, `get`
- [ ] `select`, `reject`, `merge`

**Tests:** `tests/spec/runtime/map_methods.rg`

### 13.3 String Methods

- [x] `length`, `empty?`
- [x] `include?`, `start_with?`, `end_with?`
- [x] `upcase`, `downcase`, `strip`
- [x] `split`
- [ ] `chars`, `lines`, `bytes`
- [ ] `replace`, `reverse`
- [ ] `to_i`, `to_f`

**Tests:** `tests/spec/runtime/string_methods.rg`

### 13.4 Integer Methods

- [x] `even?`, `odd?`, `zero?`, `positive?`, `negative?`
- [x] `abs`, `clamp`
- [x] `times`, `upto`, `downto`
- [x] `to_s`, `to_f`

**Tests:** `tests/spec/runtime/integer_methods.rg`, `tests/spec/blocks/times_upto_downto.rg`

### 13.5 Float Methods

- [x] `floor`, `ceil`, `round`
- [x] `zero?`, `positive?`, `negative?`
- [x] `to_i`, `to_s`

**Tests:** `tests/spec/runtime/float_methods.rg`

### 13.6 Global Functions

- [x] `puts`, `print`, `p`
- [ ] `gets`
- [ ] `exit`, `sleep`, `rand`

**Tests:** `tests/spec/runtime/global_functions.rg`

---

## Phase 14: Not Yet Implemented

These features are in the spec but marked as `skip` in tests. Implement after core features are solid.

### 14.1 Generics (Section 6)

- [ ] Generic functions `def identity<T>(x : T) -> T`
- [ ] Generic classes `class Box<T>`
- [ ] Generic interfaces
- [ ] Type constraints `T : Comparable`

**Tests:** `tests/spec/generics/*.rg`

### 14.2 Enums (Section 7)

- [ ] Basic enums `enum Status ... end`
- [ ] Explicit values `Ok = 200`
- [ ] Enum methods `to_s`, `values`, `from_string`

**Tests:** `tests/spec/enums/*.rg`

### 14.3 Structs (Section 12)

- [ ] `struct Name ... end`
- [ ] Value semantics
- [ ] Immutability
- [ ] Auto-generated constructor, equality, hash

**Tests:** `tests/spec/structs/*.rg`

### 14.4 Function Features (Section 10)

- [ ] Default parameters `def foo(x = 10)`
- [ ] Named parameters `foo(timeout: 60)`
- [ ] Variadic functions `def log(*messages)`

**Tests:** `tests/spec/functions/default_params.rg`, `named_params.rg`, `variadic.rg`

### 14.5 Destructuring (Section 9.3)

- [ ] Splat `first, *rest = items`
- [ ] Map destructuring `{name:, age:} = data`

**Tests:** `tests/spec/control_flow/destructuring.rg`

---

## Progress Tracking

Run tests to see current status:

```bash
# Summary of pass/fail/skip counts
make test-spec 2>&1 | grep -E "^\s+--- (PASS|FAIL|SKIP)" | awk '{print $2}' | cut -d: -f1 | sort | uniq -c

# Run a specific category
go test ./tests/... -run "TestSpecs/types" -v

# Run a specific test
go test ./tests/... -run "TestSpecs/types/composite_array" -v
```

When a phase is complete, all tests in that phase should PASS.
