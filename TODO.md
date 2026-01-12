# Rugby Test Coverage

Comprehensive test coverage based on spec.md.

Legend:
- [x] Tested and passing
- [ ] Not yet tested
- [N/A] Not yet implemented

---

## Missing Features

These features are defined in spec.md but not yet implemented in the compiler:

| Feature | Spec Section | Description |
|---------|--------------|-------------|
| Modules/Mixins | §8 | `module` keyword, `include` for mixing in modules, module field embedding |
| Concurrency | §13 | `go` keyword, `Chan[T]` channels, `select` statement, channel operations |
| `super` keyword | §7.7 | Calling parent implementation in inherited methods |
| Accessor macros | §7.4 | `getter`, `setter`, `property` declarative macros |

---

## 2. Compilation Model

### 2.1 Bare Scripts (Top-Level Execution)
- [x] Top-level statements execute in source order
- [x] Single file with only top-level statements generates `func main()`
- [x] `def`, `class`, `interface` at top level become package-level constructs
- [x] Error: Top-level statements + explicit `def main` in same file
- [x] Script with functions: functions lifted, calls in generated main
- [x] Script with classes: classes lifted, instantiation in generated main

## 3. Lexical Structure

### 3.1 Comments
- [x] Single line comments with `#`
- [x] Comments at end of line
- [x] Comments preserve line numbers for error reporting

### 3.2 Blocks
- [x] `do ... end` block syntax
- [x] `{ ... }` block syntax
- [x] Block with single parameter `{ |x| ... }`
- [x] Block with multiple parameters `{ |x, y| ... }`
- [x] Block with no parameters `do ... end`

### 3.3 Strings
- [x] Normal string literals `"hello"`
- [x] String interpolation `"Hello, #{name}!"`
- [x] Interpolation with expressions `"sum: #{a + b}"`
- [x] Interpolation with method calls `"len: #{items.length}"`
- [x] Empty string `""`

## 4. Types

### 4.1 Primitive Types
- [x] `Int` → `int`
- [x] `Int64` → `int64`
- [x] `Float` → `float64`
- [x] `Bool` → `bool`
- [x] `String` → `string`
- [x] `Symbol` → `string` (`:ok` → `"ok"`)
- [x] `Bytes` → `[]byte`

### 4.1.1 Symbols
- [x] Symbol syntax `:name`
- [x] Symbol with underscores `:not_found`

### 4.2 Composite Types
- [x] `Array[T]` → `[]T`
- [x] `Map[K, V]` → `map[K]V`
- [x] `T?` optionals
- [x] `Range` type

### 4.2.1 Range Type
- [x] Inclusive range `start..end`
- [x] Exclusive range `start...end`
- [x] Range in for loop
- [x] `range.each { |i| }`
- [x] `range.include?(n)` / `range.contains?(n)`
- [x] `range.to_a`
- [x] `range.size` / `range.length`
- [x] Range as first-class value

### 4.3 Type Inference
- [x] Local variable inference from assignment
- [x] Explicit type annotation `y : Int64 = 5`
- [x] Function parameters require explicit types
- [x] Instance variable inference from initialize
- [x] Instance variable from parameter promotion `def initialize(@name : String)`
- [x] Explicit field declaration `@field : Type`

### 4.4 Optionals (`T?`)

#### 4.4.1 Optional Operators
- [x] Nil coalescing `??`: `expr ?? default`
- [x] Safe navigation `&.`: `obj&.method`
- [x] Chained safe navigation `user&.address&.city`
- [x] Combined `&.` and `??`

#### 4.4.2 Optional Methods
- [x] `ok?` / `present?` returns Bool
- [x] `nil?` / `absent?` returns Bool
- [x] `unwrap!` returns T, panics if absent

#### 4.4.3 The `if let` Pattern
- [x] `if let user = find_user(id)` unwraps optional
- [x] `if let` with else branch
- [x] `if let` compiles to Go comma-ok idiom

#### 4.4.4 Tuple Unpacking
- [x] `user, ok = find_user(id)` unpacks optional
- [x] Unpacked value is non-optional type
- [x] Unpacked ok is Bool

### 4.5 The `nil` Keyword
- [x] `nil` valid for optional types
- [x] `nil` valid for `error` type
- [x] Return nil from `-> T?` function
- [x] Assign nil to optional variable
- [x] Compare error to nil

## 5. Variables & Control Flow

### 5.1 Variables & Operators
- [x] Variable declaration with `=`
- [x] Variable reassignment
- [x] Shadowing in nested scopes
- [x] Compound assignment `+=`, `-=`, `*=`, `/=`
- [x] `||=` for Bool type
- [x] `||=` for optional types

### 5.2 Conditionals
- [x] `if` with Bool condition
- [x] Error: `if` with non-Bool condition (no truthiness)
- [x] `if let` for optional unwrapping
- [x] `if let` with type assertion `obj.as(Type)`
- [x] `unless` (inverse of if)
- [x] `unless` with else
- [x] `case` expression (value switch)
- [x] `case` with multiple values `when 200, 201`
- [x] `case` with else
- [x] `case_type` (type switch)

### 5.3 Imperative Loops
- [x] `for item in items`
- [x] `for i in 0..10` (inclusive range)
- [x] `for i in 0...10` (exclusive range)
- [x] `while cond`
- [x] `break` exits loop
- [x] `next` continues to next iteration
- [x] `return` from inside loop

### 5.4 Functional Blocks
- [x] Block creates new scope
- [x] `return` in block returns value to iterator

### 5.5 Statement Modifiers
- [x] `statement if condition`
- [x] `statement unless condition`
- [x] `break if cond`
- [x] `next unless cond`
- [x] `return if cond`

## 6. Functions

### 6.1 Definition
- [x] Basic function definition
- [x] Function with typed parameters

### 6.2 Return Types
- [x] Single return type `-> Int`
- [x] Multiple return types `-> (Int, Bool)`
- [x] No return type (void)

### 6.3 Return Statements
- [x] Implicit return (last expression)
- [x] Explicit `return` for early exit
- [x] Multiple return values `return 0, false`

### 6.4 Errors
- [x] Function returning `error`
- [x] Function returning `(T, error)`

### 6.5 Calling Convention
- [x] Parentheses required for method calls
- [x] Properties don't use parentheses (getter)

## 7. Classes

### 7.1 Definition
- [x] Basic class definition
- [x] Class with `implements` interface

### 7.2 Instance Variables & Layout
- [x] Explicit field declaration
- [x] Parameter promotion in initialize
- [x] Field inference from initialize assignment

### 7.3 Initialization (`new`)
- [x] `ClassName.new(args)` calls initialize
- [x] Generated Go `NewClassName` function

### 7.4 Accessors
- [N/A] `getter name : Type` generates getter (not yet implemented)
- [N/A] `setter name : Type` generates setter (not yet implemented)
- [N/A] `property name : Type` generates both (not yet implemented)

### 7.5 Methods and Receivers
- [x] All methods use pointer receivers
- [x] Method can access instance variables
- [x] Method can modify instance variables

### 7.7 Inheritance & Specialization
- [x] `class Child < Parent` syntax (embedding)
- [x] Data layout embeds parent
- [N/A] Method specialization (cloning) - not yet implemented
- [N/A] `super` calls parent implementation - not yet implemented

### 7.9 Explicit Implementation
- [x] `class Foo implements Bar` syntax
- [x] Compile error if methods missing

### 7.10 Polymorphism & Interfaces
- [x] Interface types for polymorphism

## 8. Modules (Mixins)
- [N/A] Module definition - not yet implemented
- [N/A] `include ModuleName` - not yet implemented
- [N/A] Module fields embedded in class - not yet implemented
- [N/A] Module methods specialized into class - not yet implemented

## 9. Interfaces

### 9.1 Declaration
- [x] Interface with method signatures
- [x] Interface names CamelCase

### 9.2 Interface Inheritance
- [x] `interface IO < Reader, Writer`
- [x] Composed interface requires all methods

### 9.3 Satisfaction & `implements`
- [x] Structural typing (implicit satisfaction)
- [x] `implements` as compile-time assertion

### 9.4 The `any` Type
- [x] `any` maps to `interface{}`

### 9.5 Runtime Casting
- [x] `obj.is_a?(Interface)` returns Bool
- [x] `obj.as(Interface)` returns `Interface?`
- [x] `if let w = obj.as(Writer)`
- [x] `obj.as(Type) ?? default`
- [x] `obj.as(Type).unwrap!`

## 10. Visibility & Naming

### 10.1 Core Principle
- [x] `pub` exports to Go (uppercase)
- [x] No `pub` = internal (lowercase)

### 10.2 What Can Be `pub`
- [x] `pub def` function
- [x] `pub class`
- [x] `pub def` inside pub class
- [x] `pub interface`

### 10.3 Rugby Naming Conventions
- [x] Types in CamelCase
- [x] Functions/methods/variables in snake_case
- [x] Predicate methods end in `?`
- [x] `?` methods must return Bool

### 10.4 Go Name Generation
- [x] `snake_case` → `camelCase` (internal)
- [x] `snake_case` → `PascalCase` (pub)
- [x] Acronyms handled (`user_id` → `userID`)

## 11. Go Interop

### 11.1 Imports
- [x] `import net/http`
- [x] `import encoding/json as json`

### 11.2 Calling Go Functions
- [x] `http.Get(url)` syntax
- [x] `io.read_all(r)` → `io.ReadAll(r)`

### 11.4 Defer
- [x] `defer expr.method()` syntax
- [x] Compiles to Go defer

## 12. Runtime Package

### 12.3 Array Methods
- [x] `each { |x| }`
- [x] `each_with_index { |x, i| }`
- [x] `map { |x| }`
- [x] `select { |x| }` / `filter { |x| }`
- [x] `reject { |x| }`
- [x] `find { |x| }` / `detect { |x| }`
- [x] `any? { |x| }`
- [x] `all? { |x| }`
- [x] `none? { |x| }`
- [x] `include?(val)` / `contains?(val)`
- [x] `reduce(init) { |acc, x| }`
- [x] `sum` (numeric arrays)
- [x] `min` / `max`
- [x] `first` / `last`
- [x] `length` / `size`
- [x] `empty?`
- [x] `reverse!` / `reverse`
- [x] `sort`
- [x] `join`
- [x] `flatten`
- [x] `uniq`
- [x] `shuffle`
- [x] `sample`
- [x] `first_n` / `last_n`
- [x] `rotate`

### 12.4 Map Methods
- [x] `keys`
- [x] `values`
- [x] `length` / `size`
- [x] `has_key?(k)` / `key?(k)`
- [x] `fetch(k, default)`
- [x] `select { |k, v| }`
- [x] `reject { |k, v| }`
- [x] `merge(other)`
- [x] `delete(k)`
- [x] `clear`
- [x] `invert`

### 12.5 String Methods
- [x] `length` / `size`
- [x] `char_length`
- [x] `empty?`
- [x] `include?(sub)` / `contains?(sub)`
- [x] `upcase`
- [x] `downcase`
- [x] `capitalize`
- [x] `strip`
- [x] `lstrip` / `rstrip`
- [x] `replace(old, new)`
- [x] `reverse`
- [x] `split(sep)`
- [x] `chars`
- [x] `to_i` → `(Int, error)`
- [x] `to_f` → `(Float, error)`

### 12.6 Integer Methods
- [x] `even?`
- [x] `odd?`
- [x] `abs`
- [x] `clamp(min, max)`
- [x] `times { |i| }`
- [x] `upto(max) { |i| }`
- [x] `downto(min) { |i| }`

### 12.7 Float/Math Methods
- [x] `floor`
- [x] `ceil`
- [x] `round`
- [x] `sqrt`
- [x] `pow`

### 12.9 Global Functions (Kernel)
- [x] `puts(args...)`
- [x] `print(args...)`
- [x] `p(args...)`
- [x] `exit(code)`
- [x] `sleep(seconds)`
- [x] `rand(n)`
- [x] `rand` (no arg)

## 13. Concurrency
- [N/A] `go func_call()` - not yet implemented
- [N/A] `go do ... end` - not yet implemented
- [N/A] Channels - not yet implemented
- [N/A] `select` - not yet implemented

## 15. Errors

### 15.1 The `error` Type
- [x] `error` type maps to Go `error`
- [x] `nil` represents no error

### 15.2 Fallible Function Signatures
- [x] `-> (T, error)` signature
- [x] `-> error` signature

### 15.3 Postfix Bang Operator (`!`)
- [x] `call()!` propagates error
- [x] `!` unwraps `(T, error)` to `T`
- [x] `!` on `error`-only for control flow
- [x] Chained `!`: `a()!.b()!.c()!`

### 15.4 The `rescue` Keyword
- [x] Inline default: `call() rescue default`
- [x] Block form: `call() rescue do ... end`
- [x] Error binding: `call() rescue => err do ... end`

### 15.7 Explicit Handling
- [x] Manual `if err != nil` pattern

### 15.8 Panics
- [x] `panic "message"`

### 15.9 Error Utilities
- [x] `error_is?(err, target)`
- [x] `error_as(err, Type)` returns `Type?`

---

## Edge Cases Tested

### Parser Edge Cases
- [x] Empty function body
- [x] Empty class body
- [x] Empty interface body
- [x] Deeply nested expressions
- [x] Chained method calls
- [x] Nil coalesce chains
- [x] Safe navigation chains
- [x] Tuple unpacking
- [x] Inclusive/exclusive ranges in for loops
- [x] Case type switch
- [x] Symbol literals
- [x] Defer statements
- [x] Panic statements

### Runtime Edge Cases
- [x] Range with break
- [x] Empty range iteration
- [x] Array methods with break/continue
- [x] Map operations on empty maps
- [x] String operations on empty strings
- [x] Math functions with edge values (NaN, Inf)

---

## Not Yet Implemented (Future Work)

These features are defined in spec.md but not yet implemented in the compiler:

1. **Modules/Mixins (Section 8)**
   - `module` keyword
   - `include` for mixing in modules
   - Module field embedding
   - Method specialization from modules

2. **Concurrency (Section 13)**
   - `go` keyword for goroutines
   - `Chan[T]` channel type
   - `select` statement
   - Channel operations (`<<`, `.receive`, `.try_receive`)

3. **Inheritance Features (Section 7.7)**
   - `super` keyword for calling parent methods
   - Method specialization (cloning parent methods)

4. **Accessor Macros (Section 7.4)**
   - `getter` macro
   - `setter` macro
   - `property` macro
