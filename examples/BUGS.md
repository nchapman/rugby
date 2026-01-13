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
| 01_hello.rg | ✅ PASS | None |
| 02_types.rg | ✅ PASS | Int/Float methods now work |
| 03_control_flow.rg | ✅ PASS | None |
| 04_loops.rg | ✅ PASS | Predicate methods on arrays now work |
| 05_functions.rg | ✅ PASS | Optional return types now work |
| 06_classes.rg | ✅ PASS | Subclass constructors, method calls without parens now work |
| 07_interfaces.rg | ✅ PASS | Interface structural typing now works |
| 08_modules.rg | ❌ FAIL | Pointer printing instead of values |
| 09_blocks.rg | ✅ PASS | Block methods now work |
| 10_optionals.rg | ✅ PASS | Optional handling now works (if let, safe nav, tuple unpacking, map/each) |
| 11_errors.rg | ❌ FAIL | os.ReadFile multi-value return issues |
| 12_strings.rg | ❌ FAIL | String methods (contains?, etc.) |
| 13_ranges.rg | ❌ FAIL | Range.size method not implemented |
| 14_go_interop.rg | ✅ PASS | Multi-value Go function returns now work |
| 15_concurrency.rg | ❌ FAIL | Chan generic syntax, sync.WaitGroup.new |

---

## Fixed Bugs

### ~~BUG-001: Integer methods~~ ✅ FIXED
Integer methods (`even?`, `odd?`, `abs`, `positive?`, `negative?`, `zero?`, `to_s`, `to_f`) now work via runtime calls.

### ~~BUG-002: Float methods~~ ✅ FIXED
Float methods (`floor`, `ceil`, `round`, `truncate`, `abs`, `positive?`, `negative?`, `zero?`, `to_s`, `to_i`) now work via runtime calls.

### ~~BUG-003: Integer to_s/to_f~~ ✅ FIXED
Conversion methods now implemented.

### ~~BUG-004: `||=` operator~~ ✅ FIXED
The `||=` operator now works correctly.

### ~~BUG-005: Loop modifiers~~ ✅ FIXED
Statement modifiers (`while`/`until` after statements) now work, including with compound assignment.

### ~~BUG-006: Optional return types~~ ✅ FIXED
Can now return concrete types from optional return type functions.

### ~~BUG-007: getter/property with parameter promotion~~ ✅ FIXED
Getter accessors now work with parameter promotion.

### ~~BUG-009: Module method resolution~~ ✅ FIXED
Methods from included modules are now callable within the class.

### ~~BUG-010: Block methods with arguments~~ ✅ FIXED
`reduce`, `upto`, `downto` and other block methods with initial values now work.

### ~~BUG-011: Optional parameters~~ ✅ FIXED
Can now pass concrete types to optional parameters.

### ~~BUG-012: Go package imports~~ ✅ FIXED
Go standard library imports are now recognized.

### ~~BUG-013: rescue => err binding~~ ✅ FIXED
Error binding in rescue blocks now works.

### ~~BUG-015: Range slicing~~ ✅ FIXED
Range indexing now slices arrays and strings.

### ~~BUG-016: concurrently expression~~ ✅ FIXED
`concurrently do |scope| ... end` now works as an expression in assignment context.

### ~~BUG-008: Interface structural typing~~ ✅ FIXED
Classes now satisfy interfaces structurally. Interface methods are generated with PascalCase for Go compatibility. Type assertions (`is_a?` and `as`) work correctly on concrete types.

### ~~BUG-021: Multi-value Go function returns~~ ✅ FIXED
Underscore (`_`) is now recognized as a valid identifier, allowing patterns like `_, err = os.ReadFile(...)`. Multi-value assignments from Go functions now work correctly. Unknown types in multi-assignment are given `any` type.

### ~~BUG-018: Missing return detection for string methods~~ ✅ FIXED
The `to_s` and `message` methods now correctly generate implicit returns for the last expression when it's a string interpolation or other expression.

### ~~BUG-027: Compound assignment to instance variable~~ ✅ FIXED
Instance variable compound assignment (`@count += 1`) now works correctly, generating proper code like `c._count = c._count + 1`. The original example failure was due to field visibility issues, not the compound assignment pattern itself.

### ~~BUG-028: to_s call resolves to wrong method name~~ ✅ FIXED
Calling `to_s` on objects now correctly invokes the `String()` method. The semantic analyzer properly resolves `to_s` as a method call.

### ~~BUG-030: Setter assignment~~ ✅ FIXED
Property setter assignments like `person.email = "value"` now correctly generate setter method calls (e.g., `person.SetEmail("value")` for pub classes or `person.setEmail("value")` for private classes).

### ~~BUG-026: Method calls without parentheses~~ ✅ FIXED
Methods without parentheses are now correctly identified and called as methods (not treated as field access). The semantic analyzer properly resolves selector kinds for class methods.

### ~~BUG-029: Subclass constructors~~ ✅ FIXED
Subclasses that don't define their own `initialize` method now automatically get a constructor that delegates to the parent class constructor.

### ~~BUG-031: Class method parameter types~~ ✅ FIXED
Method parameters with class types are now correctly generated as pointer types (e.g., `other *Point` instead of `other Point`), matching Ruby/Rugby's reference semantics.

### ~~BUG-032: Embedded struct field access~~ ✅ FIXED
Instance variables accessed in subclasses now correctly use the underscore prefix from parent class accessor fields, fixing issues where `@name` would resolve to a method value instead of the actual field value.

### ~~BUG-017: Predicate methods on arrays~~ ✅ FIXED
Array methods `any?`, `empty?`, `shift`, and `pop` now work correctly. `any?` returns true if the array has any elements, `empty?` returns the opposite. `shift` removes and returns the first element, `pop` removes and returns the last element. These methods correctly modify the underlying array via pointer semantics.

### ~~BUG-020: "if let" pattern and optional handling~~ ✅ FIXED
Full optional support now works:
- `if let` pattern: binds unwrapped value when optional is present
- Safe navigation (`&.`): method chaining through optionals
- Tuple unpacking: `val, ok = optional_expr` extracts value and presence flag
- Nil coalescing (`??`): provides default values for nil optionals
- Optional methods: `ok?`, `nil?`, `present?`, `absent?`, `unwrap` work on optionals
- `map`/`each` on optionals: transforms or iterates if value is present

---

## Remaining Bugs

### BUG-019: Pointer printing instead of values
**File:** 08_modules.rg
**Error:** Prints memory addresses instead of values when printing objects

---

### BUG-022: String methods not implemented
**File:** 12_strings.rg
**Code:**
```ruby
greeting.contains?("World")
```
**Error:** `greeting.contains undefined (type string has no field or method contains)`
**Expected:** String methods like `contains?`, `starts_with?`, etc. should work

---

### BUG-023: Range.size method
**File:** 13_ranges.rg
**Code:**
```ruby
r = 1..10
puts r.size
```
**Error:** `invalid argument: r for built-in len`
**Expected:** Range should have a `size` method

---

### BUG-024: Chan generic syntax not recognized
**File:** 15_concurrency.rg
**Code:**
```ruby
ch = Chan[Int].new(3)
```
**Error:** `undefined: 'Chan'`
**Expected:** Generic channel syntax should work

---

### BUG-025: sync.WaitGroup.new syntax
**File:** 15_concurrency.rg
**Code:**
```ruby
wg = sync.WaitGroup.new
```
**Error:** `sync.WaitGroup.New undefined`
**Expected:** `.new` should map to Go's zero-value initialization or constructor

---

## Priority Order for Remaining Fixes

### Medium Priority
1. **BUG-022**: String methods
2. **BUG-023**: Range.size method

### Lower Priority
3. **BUG-019**: Pointer printing
4. **BUG-024/025**: Generic Chan syntax and sync.WaitGroup.new
