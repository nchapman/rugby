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
| 04_loops.rg | ❌ FAIL | Predicate methods on arrays (any?, empty?) |
| 05_functions.rg | ✅ PASS | Optional return types now work |
| 06_classes.rg | ❌ FAIL | "missing return" for string-returning methods |
| 07_interfaces.rg | ✅ PASS | Interface structural typing now works |
| 08_modules.rg | ❌ FAIL | Pointer printing instead of values |
| 09_blocks.rg | ✅ PASS | Block methods now work |
| 10_optionals.rg | ❌ FAIL | "if let" pattern not implemented |
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

---

## Remaining Bugs

### BUG-017: Predicate methods on arrays
**File:** 04_loops.rg
**Code:**
```ruby
items = [1, 2, 3]
puts items.shift while items.any?
```
**Error:** `items.any undefined (type []int has no field or method any)`
**Expected:** `any?` and `empty?` should work on arrays

---

### BUG-018: Missing return detection for string methods
**File:** 06_classes.rg
**Code:**
```ruby
def to_s -> String
  "(#{@x}, #{@y})"
end
```
**Error:** `missing return`
**Expected:** Last expression should be implicit return

---

### BUG-019: Pointer printing instead of values
**File:** 08_modules.rg
**Error:** Prints memory addresses instead of values when printing objects

---

### BUG-020: "if let" pattern not implemented
**File:** 10_optionals.rg
**Code:**
```ruby
if let user = find_user(2)
  puts "Found: #{user.name}"
end
```
**Error:** `assignment mismatch: 2 variables but findUser returns 1 value`
**Expected:** `if let` should bind the unwrapped value

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

### High Priority
1. **BUG-018**: Missing return detection - common pattern

### Medium Priority
2. **BUG-017**: Predicate methods on arrays
3. **BUG-020**: "if let" pattern
4. **BUG-022**: String methods
5. **BUG-023**: Range.size method

### Lower Priority
6. **BUG-019**: Pointer printing
7. **BUG-024/025**: Generic Chan syntax and sync.WaitGroup.new
