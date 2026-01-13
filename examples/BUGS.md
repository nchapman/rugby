# Compiler Bugs Found in Examples

This document tracks bugs found when testing idiomatic Rugby code from the spec against the compiler.

## Summary

| Example | Status | Bugs |
|---------|--------|------|
| 01_hello.rg | ✅ PASS | None |
| 02_types.rg | ❌ FAIL | Integer/Float methods not implemented |
| 03_control_flow.rg | ✅ PASS | None |
| 04_loops.rg | ❌ FAIL | `||=` not implemented, loop modifiers not implemented |
| 05_functions.rg | ❌ FAIL | Cannot return concrete type from optional return |
| 06_classes.rg | ❌ FAIL | getter/property with parameter promotion broken |
| 07_interfaces.rg | ❌ FAIL | Structural typing for interface parameters broken |
| 08_modules.rg | ❌ FAIL | Module method calls from within class broken |
| 09_blocks.rg | ❌ FAIL | Block methods with arguments (reduce, upto, etc) broken |
| 10_optionals.rg | ❌ FAIL | Optional return types, passing concrete to optional param |
| 11_errors.rg | ❌ FAIL | Go imports not recognized, rescue => err not implemented |
| 12_strings.rg | ❌ FAIL | Range slicing not implemented |
| 13_ranges.rg | ❌ FAIL | Range slicing not implemented |
| 14_go_interop.rg | ❌ FAIL | Go imports not recognized |
| 15_concurrency.rg | ❌ FAIL | `concurrently` not implemented |

---

## Bug Details

### BUG-001: Integer methods not implemented (spec 12.6)
**File:** 02_types.rg
**Spec:** Section 12.6
**Code:**
```ruby
x = -7
puts x.even?      # should work
puts x.odd?       # should work
puts x.abs        # should work
puts x.positive?  # should work
```
**Error:** `x.even undefined (type int has no field or method even)`
**Expected:** These methods should be inlined to Go expressions or runtime calls.

---

### BUG-002: Float methods not implemented (spec 12.7)
**File:** 02_types.rg
**Spec:** Section 12.7
**Code:**
```ruby
f = 3.7
puts f.floor   # should call math.Floor
puts f.ceil    # should call math.Ceil
puts f.round   # should call math.Round
```
**Error:** `f.floor undefined (type float64 has no field or method floor)`

---

### BUG-003: Integer to_s/to_f not implemented (spec 12.6)
**File:** 02_types.rg
**Spec:** Section 12.6
**Code:**
```ruby
num = 42
puts num.to_s  # should call strconv.Itoa
puts num.to_f  # should cast to float64
```
**Error:** `num.toS undefined`

---

### BUG-004: `||=` operator not implemented (spec 5.1)
**File:** 04_loops.rg
**Spec:** Section 5.1
**Code:**
```ruby
cache : String? = nil
cache ||= "default value"
```
**Error:** Parser fails with `expected 'end' to close function`

---

### BUG-005: Loop modifiers not implemented (spec 5.6)
**File:** 04_loops.rg
**Spec:** Section 5.6
**Code:**
```ruby
items = [1, 2, 3]
puts items.shift while items.any?

counter = 0
counter += 1 until counter == 3
```
**Error:** Parser fails - doesn't recognize `statement while condition` form

---

### BUG-006: Cannot return concrete type from optional return type (spec 4.4)
**File:** 05_functions.rg, 10_optionals.rg
**Spec:** Section 4.4
**Code:**
```ruby
def find_user(id : Int) -> String?
  return "Alice" if id == 1
  nil
end
```
**Error:** `cannot return String as String?`
**Expected:** Concrete `String` should be implicitly convertible to `String?`

---

### BUG-007: getter/property conflicts with parameter promotion
**File:** 06_classes.rg
**Spec:** Section 7.4
**Code:**
```ruby
class Person
  getter name : String
  def initialize(@name : String)
  end
end
```
**Error:** `field and method with the same name name`
**Expected:** getter should work seamlessly with parameter promotion

---

### BUG-008: Structural typing for interface parameters broken
**File:** 07_interfaces.rg
**Spec:** Section 7.6, 9
**Code:**
```ruby
def greet(s : Speaker)
  puts s.speak
end

dog = Dog.new("Rex")
greet(dog)  # Dog implements Speaker
```
**Error:** `cannot use Dog as Speaker for parameter 's' in greet`
**Expected:** Structural typing should allow Dog (which has `speak -> String`) to be passed

---

### BUG-009: Module method calls within class not resolved
**File:** 08_modules.rg
**Spec:** Section 8
**Code:**
```ruby
module Greetable
  def greet -> String
    "Hello!"
  end
end

class Greeter
  include Greetable

  def personalized_greet -> String
    "#{greet} I'm #{@name}."  # calling greet from module
  end
end
```
**Error:** `undefined: 'greet'`
**Expected:** Methods from included modules should be callable within the class

---

### BUG-010: Block methods with arguments broken (spec 12.3, 12.6)
**File:** 09_blocks.rg
**Spec:** Section 12.3, 12.6
**Code:**
```ruby
nums.reduce(0) { |acc, n| acc + n }
1.upto(3) { |i| puts i }
3.downto(1) { |i| puts i }
```
**Error:** `wrong number of arguments for 'reduce': expected 0, got 1`
**Expected:** These methods take an initial value argument

---

### BUG-011: Cannot pass concrete type to optional parameter
**File:** 10_optionals.rg
**Spec:** Section 4.4
**Code:**
```ruby
class User
  def initialize(@name : String, @address : Address?)
  end
end

addr = Address.new("NYC", "10001")
user = User.new("Charlie", addr)  # addr is Address, param is Address?
```
**Error:** `cannot use Address as Address? for parameter '@address'`

---

### BUG-012: Go package imports not recognized
**File:** 11_errors.rg, 14_go_interop.rg
**Spec:** Section 11
**Code:**
```ruby
import errors
import strings
import os

errors.New("message")
strings.ToUpper("hello")
os.read_file("/path")
```
**Error:** `undefined: 'errors'`, `undefined: 'strings'`
**Expected:** Go standard library imports should work

---

### BUG-013: `rescue => err` binding not implemented (spec 15.3)
**File:** 11_errors.rg
**Spec:** Section 15.3
**Code:**
```ruby
data = divide(5, 0) rescue => err do
  puts "Error: #{err}"
  -999
end
```
**Expected:** Error should be captured in `err` variable

---

### BUG-014: `error_is?` and `error_as` not implemented (spec 15.7)
**File:** 11_errors.rg
**Spec:** Section 15.7
**Code:**
```ruby
if error_is?(err, os.ErrNotExist)
  puts "File not found"
end
```

---

### BUG-015: Range slicing not implemented (spec 4.2)
**File:** 12_strings.rg, 13_ranges.rg
**Spec:** Section 4.2
**Code:**
```ruby
word = "Rugby"
puts word[0..2]      # should return "Rug"

letters = ["a", "b", "c", "d", "e"]
puts letters[1..3]   # should return ["b", "c", "d"]
```
**Error:** `array/string index must be Int, got Range`
**Expected:** Range indexing should slice arrays and strings

---

### BUG-016: `concurrently` not implemented (spec 13.5)
**File:** 15_concurrency.rg
**Spec:** Section 13.5
**Code:**
```ruby
result = concurrently do |scope|
  t1 = scope.spawn { compute(10) }
  await t1
end
```
**Error:** `block can only follow a method call`

---

## Priority Order for Fixes

### High Priority (Breaks fundamental features)
1. **BUG-006**: Optional return types - fundamental to Rugby's error handling
2. **BUG-011**: Optional parameters - fundamental to optional types
3. **BUG-008**: Interface structural typing - core OOP feature
4. **BUG-012**: Go imports - essential for Go interop
5. **BUG-015**: Range slicing - common Ruby idiom

### Medium Priority (Breaks common patterns)
6. **BUG-007**: getter/property with parameter promotion
7. **BUG-009**: Module method resolution
8. **BUG-010**: Block methods with arguments (reduce, upto, downto)
9. **BUG-001/002/003**: Numeric methods

### Lower Priority (Advanced features)
10. **BUG-004**: `||=` operator
11. **BUG-005**: Loop modifiers
12. **BUG-013/014**: Advanced error handling
13. **BUG-016**: Structured concurrency
