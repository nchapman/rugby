# Compiler Bugs Found in Examples

This document tracks bugs found when testing idiomatic Rugby code from the spec against the compiler.

> **Important:** The example files are our **golden set** - they represent idiomatic Rugby code
> as defined by the language specification. These files should **never be modified** just to make
> them pass. If an example fails, that's a compiler bug that needs to be fixed, not an example
> that needs to be simplified. The examples define what the language *should* do, and the
> compiler must be fixed to match.

## Summary

| Example | Status | Notes |
|---------|--------|-------|
| 01_hello.rg | PASS | None |
| 02_types.rg | FAIL | Inline type annotations, empty typed arrays |
| 03_control_flow.rg | PASS | None |
| 04_loops.rg | PASS | Predicate methods on arrays now work |
| 05_functions.rg | PASS | Optional return types now work |
| 06_classes.rg | FAIL | Self return, method chaining |
| 07_interfaces.rg | PASS | Interface structural typing now works |
| 08_modules.rg | PASS | Modules generate interfaces for lint compliance |
| 09_blocks.rg | FAIL | Method chaining with newlines |
| 10_optionals.rg | FAIL | Safe navigation on getters, unwrap |
| 11_errors.rg | PASS | errors.new now works |
| 12_strings.rg | PASS | String methods now work |
| 13_ranges.rg | PASS | Range methods on variables now work |
| 14_go_interop.rg | PASS | Go interop now works |
| 15_concurrency.rg | PASS | Concurrency features now work |
| fizzbuzz.rg | PASS | None |
| field_inference.rg | PASS | None |
| json_simple.rg | PASS | None |
| worker_pool.rg | FAIL | Getter access on channel receive results |
| todo_app.rg | FAIL | Map literal parsing in method body |
| http_json.rg | FAIL | resp.json method, interface indexing |

---

## Active Bugs

### BUG-036: Inline type annotation syntax
**File:** 02_types.rg

Empty typed array literals with inline annotation syntax fail:
```ruby
empty_nums : Array[Int] = []        # type mismatch: expected Array[Int], got Array[any]
empty_strs = [] : Array[String]     # undefined: 'Array'
```

### BUG-037: Self return and method chaining
**File:** 06_classes.rg

Methods that return `self` for chaining don't work correctly:
```ruby
def inc
  @count += 1
  self  # return self for chaining
end

counter.inc.inc.inc  # counter.Inc() (no value) used as value
```

### BUG-038: Method chaining with newlines
**File:** 09_blocks.rg

Method chaining that spans multiple lines fails to parse:
```ruby
result = [1, 2, 3, 4, 5, 6]
  .select { |n| n.even? }    # expected 'end' to close select
  .map { |n| n * 10 }
```

The parser sees `.select` and thinks it's a `select` statement.

### BUG-039: Safe navigation on getter methods
**File:** 10_optionals.rg

Safe navigation (`&.`) on properties defined with `getter` fails:
```ruby
city1 = user_with_addr.address&.city ?? "Unknown"
# invalid operation: cannot call user_with_addr.address() (value of type *Address)
```

### BUG-040: unwrap on optionals
**File:** 10_optionals.rg

The `unwrap` method on optionals doesn't work correctly:
```ruby
known = find_user(1).unwrap
# invalid operation: _nc5 != nil (mismatched types OptionalResult and untyped nil)
```

### BUG-050: Getter access on channel receive results
**File:** worker_pool.rg

When receiving from a channel, the result type is not tracked by the semantic analyzer, so getter access fails:
```ruby
results = Chan[Result].new(10)
result = results.receive
puts "#{result.job_id}"  # job_id is printed as method value, not called
```
The generated code is `result.jobID` instead of `result.jobID()`.

### BUG-046: Map literal in method body
**File:** todo_app.rg

Map literals with symbol shorthand in method bodies fail to parse:
```ruby
def to_map -> Map[String, any]
  {
    id: @id,           # expected key in map literal
    title: @title,     # expected type after ':'
    # ...
  }
end
```

### BUG-047: HTTP response methods
**File:** http_json.rg

HTTP response object missing expected methods:
```ruby
resp.json       # has no field or method json
resp.jsonArray  # has no field or method jsonArray
```

### BUG-048: Interface indexing
**File:** http_json.rg

Cannot index values of interface type `any`:
```ruby
post["title"]  # cannot index post (variable of interface type any)
```

---

## Fixed Bugs

### ~~BUG-049: Unused module methods (lint)~~ FIXED
Modules now generate Go interfaces and classes that include them have interface compliance checks (`var _ ModuleName = (*ClassName)(nil)`). This suppresses "unused method" lint warnings while providing compile-time verification that included methods are implemented correctly.

### ~~BUG-001: Integer methods~~ FIXED
Integer methods (`even?`, `odd?`, `abs`, `positive?`, `negative?`, `zero?`, `to_s`, `to_f`) now work via runtime calls.

### ~~BUG-002: Float methods~~ FIXED
Float methods (`floor`, `ceil`, `round`, `truncate`, `abs`, `positive?`, `negative?`, `zero?`, `to_s`, `to_i`) now work via runtime calls.

### ~~BUG-003: Integer to_s/to_f~~ FIXED
Conversion methods now implemented.

### ~~BUG-004: `||=` operator~~ FIXED
The `||=` operator now works correctly.

### ~~BUG-005: Loop modifiers~~ FIXED
Statement modifiers (`while`/`until` after statements) now work, including with compound assignment.

### ~~BUG-006: Optional return types~~ FIXED
Can now return concrete types from optional return type functions.

### ~~BUG-007: getter/property with parameter promotion~~ FIXED
Getter accessors now work with parameter promotion.

### ~~BUG-009: Module method resolution~~ FIXED
Methods from included modules are now callable within the class.

### ~~BUG-010: Block methods with arguments~~ FIXED
`reduce`, `upto`, `downto` and other block methods with initial values now work.

### ~~BUG-011: Optional parameters~~ FIXED
Can now pass concrete types to optional parameters.

### ~~BUG-012: Go package imports~~ FIXED
Go standard library imports are now recognized.

### ~~BUG-013: rescue => err binding~~ FIXED
Error binding in rescue blocks now works.

### ~~BUG-015: Range slicing~~ FIXED
Range indexing now slices arrays and strings.

### ~~BUG-016: concurrently expression~~ FIXED
`concurrently do |scope| ... end` now works as an expression in assignment context.

### ~~BUG-008: Interface structural typing~~ FIXED
Classes now satisfy interfaces structurally. Interface methods are generated with PascalCase for Go compatibility. Type assertions (`is_a?` and `as`) work correctly on concrete types.

### ~~BUG-021: Multi-value Go function returns~~ FIXED
Underscore (`_`) is now recognized as a valid identifier, allowing patterns like `_, err = os.ReadFile(...)`. Multi-value assignments from Go functions now work correctly. Unknown types in multi-assignment are given `any` type.

### ~~BUG-018: Missing return detection for string methods~~ FIXED
The `to_s` and `message` methods now correctly generate implicit returns for the last expression when it's a string interpolation or other expression.

### ~~BUG-027: Compound assignment to instance variable~~ FIXED
Instance variable compound assignment (`@count += 1`) now works correctly, generating proper code like `c._count = c._count + 1`. The original example failure was due to field visibility issues, not the compound assignment pattern itself.

### ~~BUG-028: to_s call resolves to wrong method name~~ FIXED
Calling `to_s` on objects now correctly invokes the `String()` method. The semantic analyzer properly resolves `to_s` as a method call.

### ~~BUG-030: Setter assignment~~ FIXED
Property setter assignments like `person.email = "value"` now correctly generate setter method calls (e.g., `person.SetEmail("value")` for pub classes or `person.setEmail("value")` for private classes).

### ~~BUG-026: Method calls without parentheses~~ FIXED
Methods without parentheses are now correctly identified and called as methods (not treated as field access). The semantic analyzer properly resolves selector kinds for class methods.

### ~~BUG-029: Subclass constructors~~ FIXED
Subclasses that don't define their own `initialize` method now automatically get a constructor that delegates to the parent class constructor.

### ~~BUG-031: Class method parameter types~~ FIXED
Method parameters with class types are now correctly generated as pointer types (e.g., `other *Point` instead of `other Point`), matching Ruby/Rugby's reference semantics.

### ~~BUG-032: Embedded struct field access~~ FIXED
Instance variables accessed in subclasses now correctly use the underscore prefix from parent class accessor fields, fixing issues where `@name` would resolve to a method value instead of the actual field value.

### ~~BUG-017: Predicate methods on arrays~~ FIXED
Array methods `any?`, `empty?`, `shift`, and `pop` now work correctly. `any?` returns true if the array has any elements, `empty?` returns the opposite. `shift` removes and returns the first element, `pop` removes and returns the last element. These methods correctly modify the underlying array via pointer semantics.

### ~~BUG-020: "if let" pattern and optional handling~~ FIXED
Full optional support now works:
- `if let` pattern: binds unwrapped value when optional is present
- Safe navigation (`&.`): method chaining through optionals
- Tuple unpacking: `val, ok = optional_expr` extracts value and presence flag
- Nil coalescing (`??`): provides default values for nil optionals
- Optional methods: `ok?`, `nil?`, `present?`, `absent?`, `unwrap` work on optionals
- `map`/`each` on optionals: transforms or iterates if value is present

### ~~BUG-022: String methods~~ FIXED
String methods (`contains?`, `start_with?`, `end_with?`, `empty?`, `replace`, `lines`) now work correctly. Added runtime functions and stdLib table entries for all methods. Fixed type normalization in genCallExpr to properly match Go type names to Rugby type names.

### ~~BUG-023: Range methods on variables~~ FIXED
Range methods (`size`, `length`, `include?`, `contains?`, `to_a`) now work on Range variables, not just literal ranges. The codegen now checks for Range type using type inference, not just AST node type.

### ~~BUG-019: Module methods as instance methods~~ FIXED
Methods from included modules are now properly recognized as class methods. The semantic analyzer adds module methods to `cls.Methods` so that `GetMethod` finds them and generates proper method calls instead of method value references.

### ~~BUG-024: Chan generic syntax~~ FIXED
Generic channel syntax `Chan[Int].new(3)` now works. The codegen handles `Chan[Type].new(capacity)` and generates `make(chan type, capacity)`.

### ~~BUG-025: sync.WaitGroup.new syntax~~ FIXED
Go package type `.new` syntax like `sync.WaitGroup.new` now works, generating zero-value initialization `&sync.WaitGroup{}`.

### ~~BUG-034: Await returns any type, arithmetic fails~~ FIXED
The `await` keyword now properly types its result as `any` when the task element type is unknown. The codegen uses `runtime.Add` for arithmetic on `any`-typed values from await, enabling patterns like `r1 + r2 + r3` in concurrently blocks.

### ~~BUG-035: Unused variables in multi-value assignments~~ FIXED
Variables declared in multi-value assignments that are never used are now automatically replaced with `_` in generated code. The semantic analyzer tracks variable usage, and codegen queries this to detect and replace unused variables, avoiding Go's "declared and not used" compile error.

### ~~BUG-033: String positive indexing returns byte instead of character~~ FIXED
String indexing like `word[0]` now correctly returns characters instead of byte values. The codegen detects when the receiver is a string type and uses `runtime.AtIndex` for all string indexing, ensuring consistent behavior for both positive and negative indices.

### ~~BUG-042: Go package methods with snake_case~~ FIXED
Go package methods with snake_case names (like `strings.split`, `strings.join`) now correctly map to CamelCase Go methods. The codegen's `genStdLibPropertyMethod` and `uniqueMethods` lookup now check `isGoInterop` first to avoid incorrectly routing Go package calls through the runtime.

### ~~BUG-043: Functions defined after main not found~~ FIXED
Functions defined after `def main` are now found during semantic analysis. The analyzer checks `a.functions` map as a fallback when a symbol isn't in scope, enabling forward references to functions.

### ~~BUG-044: sync.WaitGroup snake_case methods~~ FIXED
WaitGroup and other Go type methods now map snake_case to PascalCase. The codegen tracks variables holding Go interop types (via `goInteropVars`) and uses PascalCase for method calls on those variables.

### ~~BUG-041: errors.new identifier mangling~~ FIXED
Calling `errors.new(...)` now correctly generates `errors.New(...)`. The codegen was treating `errors` as a Rugby class name and generating `newerrors(...)`. Fixed by checking if the identifier is a Go import before applying the Rugby class constructor pattern.

### ~~BUG-045: Chan type in function parameters~~ FIXED
`Chan[T]` in function parameters, return types, and channel creation now works correctly. The `mapParamType` function recursively converts class types to pointers in generic types. For loop iteration over channels now correctly generates single-variable syntax (`for x := range ch` instead of `for _, x := range ch`).

---

## Statistics

- **Passing:** 13 examples
- **Failing:** 8 examples
- **Active bugs:** 11
- **Fixed bugs:** 40
- **Total examples:** 21
