# Rugby Language Specification

**Goal:** Provide a robust, Ruby-inspired surface syntax that compiles to idiomatic Go, enabling developers to build high-performance systems with Go's power and Ruby's elegance.

Rugby is **compiled**, **statically typed**, and **deterministic**.

---

## 1. Non-Goals

Rugby **must not** support:

* Reopening classes / monkey patching
* `eval`, runtime codegen, reflection-driven magic
* `method_missing`, dynamic dispatch hacks
* Runtime modification of methods, fields, or modules
* Implicit global mutation beyond normal variables

---

## 2. Compilation Model

* Rugby compiles to **Go source** (one or more `.go` files), then uses `go build`
* Output should be **idiomatic** Go:
  * structs + methods
  * interfaces
  * package-level functions
  * standard error handling patterns
* The compiler may generate helper functions/types (prelude), but should keep them small and transparent
* Compilation must be deterministic

---

## 2.1 Bare Scripts (Top-Level Execution)

Rugby files can contain executable statements at the top level without an explicit `main` function, similar to Ruby scripts.

### Rules

1. **Top-level statements**: Any statement outside a `def`, `class`, or `interface` is considered executable top-level code.
2. **Order preserved**: Top-level statements execute in source order.
3. **Single Entry**: When compiling a directory, only **one** file may contain top-level statements (the entry point). Multiple files with top-level statements is a compile error.
4. **Definitions lifted**: `def`, `class`, and `interface` at top level become package-level Go constructs (not executed as statements).
5. **Implicit main**: If top-level statements exist and no `def main` is defined, the compiler generates `func main()` containing them.
6. **Explicit main wins**: If `def main` exists, top-level statements are a compile error (avoids ambiguity).

### Examples

**Simple script:**
```ruby
# hello.rg
puts "Hello, world!"
```

Compiles to:
```go
package main

import "rugby/runtime"

func main() {
    runtime.Puts("Hello, world!")
}
```

**Script with functions:**
```ruby
# greet.rg
def greet(name : String)
  puts "Hello, #{name}!"
end

greet("Rugby")
greet("World")
```

Compiles to:
```go
package main

import "rugby/runtime"

func greet(name string) {
    runtime.Puts(fmt.Sprintf("Hello, %s!", name))
}

func main() {
    greet("Rugby")
    greet("World")
}
```

**Script with classes:**
```ruby
# counter.rg
class Counter
  def initialize(n : Int)
    @n = n
  end

  def inc
    @n += 1
  end

  def value -> Int
    @n
  end
end

c = Counter.new(0)
c.inc()
c.inc()
puts c.value
```

### Compile Error Case

```ruby
# error: can't mix top-level statements with def main
def main
  puts "in main"
end

puts "top-level"  # ERROR: top-level statement with explicit main
```

### Package Mode vs Script Mode

| File contains | Mode | Output |
|---------------|------|--------|
| Only defs/classes | Library | `package <name>`, no main |
| Top-level statements | Script | `package main` + generated `func main()` |
| `def main` | Explicit | `package main` + user's `func main()` |
| Both top-level + `def main` | Error | Compile error |

---

## 3. Lexical Structure

### 3.1 Comments

```ruby
# single line comment
```

### 3.2 Blocks

```ruby
do ... end   # preferred
{ ... }      # alternative
```

### 3.3 Strings

```ruby
"normal string"
"interpolation: #{name}"  # compiles to fmt.Sprintf
```

String interpolation rules:
* Any expression is allowed inside `#{}`
* Always compiles to `fmt.Sprintf` for consistency
* Examples:
  * `"count: #{items.length}"`
  * `"sum: #{a + b}"`
  * `"user: #{user}"` (calls `String()` if implemented, else `%v`)

---

## 4. Types

Rugby is statically typed with inference.

### 4.1 Primitive types

| Rugby    | Go        |
|----------|-----------|
| `Int`    | `int`     |
| `Int64`  | `int64`   |
| `Float`  | `float64` |
| `Bool`   | `bool`    |
| `String` | `string`  |
| `Symbol` | `string` (interned const behavior) |
| `Bytes`  | `[]byte`  |

### 4.1.1 Symbols

Symbols are lightweight identifiers starting with `:`.

**Syntax:** `:status`, `:ok`, `:not_found`

**Compilation:**
Symbols compile to **Go strings**. `:ok` → `"ok"`

---

### 4.2 Composite types

| Rugby        | Go          |
|--------------|-------------|
| `Array[T]`   | `[]T`       |
| `Map[K, V]`  | `map[K]V`   |
| `T?`         | `(T, bool)` |
| `Range`      | `struct` (internal) |

### 4.2.1 Range Type

Ranges are first-class values representing a sequence of integers.

**Syntax:**
* `start..end` → inclusive (0..5 includes 0, 1, 2, 3, 4, 5)
* `start...end` → exclusive (0...5 includes 0, 1, 2, 3, 4)

**Compilation:** Compiles to `runtime.Range` struct:

```go
type Range struct {
    Start, End int
    Exclusive  bool
}
```

**Methods:**
* `each { |i| }` → `runtime.RangeEach(r, fn)` - iterate over values
* `include?(n)` / `contains?(n)` → `runtime.RangeContains(r, n)` - membership test
* `to_a` → `runtime.RangeToArray(r)` - materialize to `[]int`
* `size` / `length` → `runtime.RangeSize(r)` - count of elements

**Use cases:**
```ruby
r = 1..10           # store as value
for i in 0..5       # loop iteration
  puts i
end
(1..100).include?(50)  # membership test
nums = (1..5).to_a     # [1, 2, 3, 4, 5]
```

### 4.3 Type Inference

Rugby uses a flow-sensitive type inference algorithm inspired by Crystal, designed to map cleanly to Go's static structures.

**Local Variables:**
Types are inferred from their first assignment.
```ruby
x = 5           # x is Int
y : Int64 = 5   # y is Int64
```

**Function Parameters:**
Parameters **must** be explicitly typed. Use `any` for dynamic types (empty interface).
```ruby
def add(a : Int, b : Int)  # a, b are Int
def log(msg : any)         # msg is any
```

**Instance Variables:**
Instance variable types must be deducible at compile-time to generate the underlying Go struct. They are resolved from:
1.  **Explicit Declarations:** `@field : Type` at the class level.
2.  **Initialize Inference:** Assignments within the `initialize` method.
3.  **Parameter Promotion:** `def initialize(@field : Type)` syntax.

```ruby
class User
  @role : String                # 1. explicit declaration

  def initialize(@name : String) # 3. parameter promotion → @name : String
    @age = 0                     # 2. initialize inference → @age : Int
  end
end
```

New instance variables may only be introduced in `initialize` (or via class-level declarations). Introducing a new `@field` outside `initialize` is a compile-time error.

**Definite Assignment:** All non-optional fields must be definitely assigned on all control-flow paths in `initialize`. Optional fields (`T?`) may be left unassigned.

```ruby
class Example
  def initialize(flag : Bool)
    if flag
      @value = 1
    end
    # ERROR: @value not assigned when flag is false
  end
end

class ExampleFixed
  @value : Int?  # optional, may be unassigned

  def initialize(flag : Bool)
    if flag
      @value = 1
    end
    # OK: @value is optional
  end
end
```

### 4.4 Optionals (`T?`)

**Syntax:** `T?` is the universal syntax for "nullable" or "optional" values.

**Representation:**

To ensure deterministic behavior and Go-idiomatic ergonomics, optionals are unified:

1.  **Return Values:** Compiles to `(T, bool)` (standard Go "comma-ok" idiom).
2.  **Storage (Fields, Arrays, Maps):** Compiles to `runtime.Option[T]` struct:
    ```go
    type Option[T any] struct {
        Value T
        Ok    bool
    }
    ```

**Usage:**

Rugby requires explicit boolean checks. "Truthiness" of optionals is **not** supported (see 5.2).

*   **Check:** `if val.ok?` (or `present?`)
*   **Unwrap:** Use `if let` pattern.

**Assignment with check (`if let`):**

```ruby
if let n = s.to_i?
  puts n
end
```
Compiles to `if n, ok := runtime.StringToInt(s); ok { ... }`

---

## 5. Variables & Control Flow

### 5.1 Variables & Operators

* `x = expr` declares (if new) or assigns (if existing)
* Shadowing allowed in nested scopes

**Compound Assignment:**

Rugby supports compound assignment operators for arithmetic operations:

```ruby
x += 5   # x = x + 5
x -= 3   # x = x - 3
x *= 2   # x = x * 2
x /= 4   # x = x / 4
```

These compile directly to Go's equivalent operations.

**Logical OR Assignment (`||=`):**

Rugby supports `||=` for safe default assignment.

```ruby
x ||= y
```

Semantics depend on the type of `x`:
*   **Bool:** `x = x || y`
*   **Reference Types (`User`, `Array`, etc.):** `if x == nil { x = y }`
*   **Optionals (`T?`):** If value is missing (or nil), assign `y`.

*Note: For non-nullable value types (like `String` or `Int`), `||=` is generally not applicable as they cannot be nil/false (except `Bool`).*

### 5.2 Conditionals

Conditionals in Rugby are **strict**. Only `Bool` values are allowed in `if`, `unless`, and `while` conditions. There is no implicit "truthiness" for objects, numbers, or strings.

```ruby
if valid?     # OK (returns Bool)
if x > 0      # OK (returns Bool)
if user       # ERROR: User is not Bool
if user.ok?   # OK (explicit check)
```

**Unwrapping (`if let`):**

Use `if let` to handle optionals and type assertions safeley.

```ruby
if let user = find_user(id)
  puts user.name
end

if let w = obj.as?(Writer)
  w.write("hello")
end
```

**Unless:**

Rugby also supports `unless` (inverse of `if`).

```ruby
unless valid?
  puts "invalid"
else
  puts "valid"
end
```

**Case Expressions (Value Switch):**

Matches values using `==`.

```ruby
case status
when 200, 201
  puts "ok"
else
  puts "error"
end
```

**Type Switch:**

Use `case_type` (or `case type`) to match types.

```ruby
case_type x
when String
  puts "string"
when Int
  puts "int"
end
```

### 5.3 Imperative Loops (Statements)

Use loops when you need **control flow** (searching, early exit, side effects). These are statements, not expressions (they do not return a value).

**For Loop:**
The primary way to iterate with control flow.

```ruby
for item in items
  return item if item.id == 5
end
```

Compiles to: `for _, item := range items { ... }`

**Range iteration** (requires Range type - see 4.2.1):
```ruby
for i in 0..10
  puts i
end
```

Compiles to: `for i := 0; i <= 10; i++ { ... }`

**While Loop:**

```ruby
while cond
  ...
end
```

Compiles to: `for cond { ... }`

**Control Flow Keywords:**
* `break`: Exits the loop immediately.
* `next`: Skips to the next iteration (Go `continue`).
* `return`: Returns from the **enclosing function**.

### 5.4 Functional Blocks (Expressions)

Use blocks when you need to **transform data** or chain operations. Blocks in Rugby are strictly **anonymous pure functions** (lambdas).

**Syntax:**
* `do ... end`: Preferred for multi-line blocks.
* `{ ... }`: Preferred for single-line blocks.

**Semantics:**
* **Scope:** Creates a new local scope (Go function literal).
* **Return:** `return` exits the **block only** (returns a value to the iterator).
* **No Control Flow:** `break` and `next` are **not supported** inside blocks. Use iterator methods like `take_while` or `find` for early exit.

```ruby
names = users.map do |u|
  u.name
end
```

**Compilation:**
Blocks compile directly to Go function literals passed to runtime helpers.

```ruby
items.each { |x| puts x }
```

Compiles to:
```go
runtime.Each(items, func(x any) {
    runtime.Puts(x)
})
```

### 5.5 Statement Modifiers

Rugby supports suffix `if` and `unless` modifiers for concise control flow.

**Syntax:**
```ruby
<statement> if <condition>
<statement> unless <condition>
```

**Supported statements:**
* `break if cond`
* `next unless cond`
* `return if cond`
* `puts x unless cond` (any expression statement)

**Compilation:**
Lowered to a standard Go `if` block.
```ruby
break if x == 2
```
Compiles to:
```go
if x == 2 {
    break
}
```

```ruby
puts "error" unless valid?
```
Compiles to:
```go
if !valid() {
    runtime.Puts("error")
}
```

---

## 6. Functions

### 6.1 Definition

```ruby
def add(a, b)
  a + b
end
```

Compiles to `func add(a T, b T) T { ... }` with inferred types.

### 6.2 Return types

Single return:

```ruby
def add(a, b) -> Int
  a + b
end
```

Multiple returns (Go-style):

```ruby
def parse_int(s) -> (Int, Bool)
  ...
end
```

### 6.3 Return statements

Rugby supports both explicit and implicit returns:

**Implicit return:** The last expression in a function is returned:

```ruby
def add(a, b)
  a + b  # returned
end
```

**Explicit return:** Use `return` for early exit or clarity:

```ruby
def find_user(id) -> User?
  return nil if id < 0  # early return
  users[id]             # implicit return
end
```

Multiple return values:

```ruby
def parse(s) -> (Int, Bool)
  return 0, false if s.empty?
  s.to_i, true
end
```

### 6.4 Errors

Rugby adopts Go error patterns. Surface `error` directly:

```ruby
def read(path : String) -> (Bytes, error)
  os.read_file(path)
end
```

### 6.5 Calling Convention

To avoid ambiguity, Rugby requires parentheses for method calls that are not property accessors.

*   **Valid:** `inc()`, `add(1, 2)`, `puts("hi")`
*   **Invalid:** `inc` (ambiguous with function value), `add 1, 2` (except in top-level scripts where relaxed parsing may apply).
*   **Properties:** `user.age` (getter) and `user.age = 5` (setter) do not use parentheses.

---

## 7. Classes

Classes in Rugby describe the shape of objects (Go structs) and their behavior (methods).

### 7.1 Definition

```ruby
class User implements Identifiable
  def initialize(@name : String, email : String, @age : Int)
    @email = email
  end
end
```

### 7.2 Instance Variables & Layout (Strict Inference)

Rugby classes compile directly to Go structs. Instance variables may be inferred; explicit declarations are optional.

**Rules:**

New instance variables may be introduced only by:
1. **Explicit Declaration:** `@field : Type` at the class level.
2. **Parameter Promotion:** `def initialize(@field : Type)` syntax.
3. **First Assignment in Initialize:** `@field = expr` within `initialize`.

Introducing a new instance variable outside `initialize` is a **compile-time error**.

The compiler determines the complete struct layout at compile time.

**Definite Assignment:** All non-optional fields must be definitely assigned on all control-flow paths in `initialize`. Optional fields (`T?`) may be left unassigned.

```ruby
class Point
  def initialize(@x : Int, @y : Int)
  end
end

class User
  @email : String  # explicit declaration (optional)

  def initialize(@name : String, @age : Int)
    @email = "unknown"
  end

  def birthday
    @age += 1  # OK: uses existing field
    @new = 5   # ERROR: cannot introduce field outside initialize
  end
end
```

### 7.3 Initialization (`new`)

Rugby separates allocation from initialization, but exposes a unified `new` class method.

*   `User.new(...)` is automatically generated.
*   It allocates the struct (Go zero value).
*   It calls `initialize(...)` on the new instance.

```ruby
# Source
u = User.new("Alice", "a@b.com", 30)

# Generated Go
func NewUser(name string, email string, age int) *User {
    u := &User{}
    u.Initialize(name, email, age)
    return u
}
```

### 7.4 Accessors (Properties)

Rugby provides macros to reduce boilerplate, inspired by Crystal. Each macro **declares the field** and generates accessor methods:

*   `getter name : String` → Declares `@name : String` + generates `def name -> String`
*   `setter name : String` → Declares `@name : String` + generates `def name=(v : String)`
*   `property name : String` → Declares `@name : String` + generates both methods

You do **not** need to separately declare `@name : String` when using these macros. Writing both is redundant.

```ruby
class User
  property name : String   # declares @name and generates getter/setter
  getter age : Int         # declares @age and generates getter only

  def initialize(@name : String, @age : Int)
  end
end
```

These compile to idiomatic Go methods:
*   `getter` → `func (r *T) Name() T`
*   `setter` → `func (r *T) SetName(v T)` (Standard Go setter pattern)

### 7.5 Methods and receivers

**Rule:** All methods use **pointer receivers** (`func (r *T)`) by default.
* This ensures Ruby-like reference semantics (modifying `@field` always works).
* Prevents accidental mutation of struct copies.

```ruby
class User
  def name
    @name
  end
end
```

Compiles to: `func (u *User) name() string { ... }`

### 7.7 Inheritance & Specialization

Rugby supports code reuse through **Specialization**. While it uses Go's struct embedding for data layout, it ensures Ruby-like method dispatch by specializing inherited methods.

```ruby
class Parent
  def hello
    puts "Hello, #{name}" # Implicitly self.name()
  end
  def name; "Parent"; end
end

class Child < Parent
  def name; "Child"; end
end

Child.new.hello # Output: "Hello, Child"
```

**Semantics:**
*   **Single Inheritance:** A class may inherit from only **one** parent class (`class Child < Parent`).
*   **Data Layout:** `class Child < Parent` embeds the `Parent` struct into `Child`.
*   **Method Specialization:** The compiler automatically "clones" methods from `Parent` into `Child`.
    *   **Dispatch:** Inside a specialized method, implicit calls (`foo()`) are treated as `self.foo()`. They dispatch against the **Child's** method set.
    *   **Receiver Update:** The receiver (`self`) is updated from `*Parent` to `*Child`.

**Super:**
The `super` keyword statically binds to the **Parent's original implementation**. It does not call the specialized clone, preventing recursion.

```ruby
class Child < Parent
  def hello
    super # Calls Parent.hello(*Parent) embedded field
  end
end
```

**Strict Typing (No Subtype Polymorphism):**
Concrete types are **invariant**. A variable of type `Parent` cannot hold a `Child`.
*   Invalid: `p : Parent = Child.new()`
*   Valid: `p : Interface = Child.new()` (See 7.8)

**Important:** This mechanism is for **composition and code reuse**, not traditional OOP inheritance. Use interfaces for polymorphism.

**Performance Note:**
Specialization increases binary size (code duplication) and compile time but ensures **zero runtime overhead** for method calls (no vtable lookups).

### 7.9 Explicit Implementation (Optional)

While Rugby uses structural typing (implicit satisfaction), classes may explicitly declare conformance to an interface using `implements`. This serves as documentation and a compile-time assertion.

```ruby
interface Runnable
  def run
end

class Job implements Runnable
  def run
    puts "working"
  end
end
```

**Compilation:**
The compiler verifies that `Job` satisfies `Runnable`. If a method is missing or signatures don't match, compilation fails with a descriptive error. This produces no runtime code overhead.

### 7.10 Polymorphism & Interfaces

Rugby uses **Interfaces** for polymorphism. Since concrete inheritance is for code reuse (not subtyping), you must use Interfaces to treat different classes uniformly.

---

## 8. Modules (Mixins)

Modules are the primary mechanism for sharing behavior across classes.

### 8.1 Definition

```ruby
module Callable
  def call
    puts "calling..."
  end
end

module Named
  property name : String
end
```

### 8.2 Including Modules

```ruby
class Worker
  include Callable
  include Named
  
  def work
    call()        # Calls specialized Callable#call
    self.name = "Job" 
  end
end
```

**Semantics:**
*   **Composition:** `include M` embeds the Module's fields into the host class.
*   **Specialization:** Methods are specialized (cloned) into the host class, allowing them to call host methods.
*   **Collisions:** If two modules (or a module and parent) define the same method/field, it is a **compile-time error** unless the host class explicitly overrides it.


---

## 9. Interfaces

Interfaces define **capabilities**. They are pure contracts describing a set of behaviors (methods) without implementation or state.

### 9.1 Declaration

Interfaces are defined using the `interface` keyword and contain one or more method signatures.

```ruby
interface Speaker
  def speak -> String
end

interface Calculator
  def add(a : Int, b : Int) -> Int
end
```

**Syntax Rules:**
*   Methods in an interface **cannot** have a body.
*   Method signatures follow the same rules as `def` (parameters and return types).
*   Interface names must be `CamelCase`.

**Go Compilation:**
Interface methods are **implicitly exported** (PascalCase in Go) to allow for cross-package satisfaction.
```go
type Speaker interface {
    Speak() string
}
```

### 9.2 Interface Inheritance (Composition)

Interfaces can inherit from (embed) other interfaces using the `<` operator.

```ruby
interface Reader
  def read -> String
end

interface Writer
  def write(data : String)
end

# IO requires methods from both Reader and Writer, plus its own
interface IO < Reader, Writer
  def close
end
```

**Semantics:**
*   A type satisfies `IO` only if it implements `read`, `write`, and `close`.
*   This compiles directly to **Interface Embedding** in Go.

### 9.3 Satisfaction & `implements`

Rugby uses **Structural Typing** (like Go). A class satisfies an interface automatically if it has the required methods. 

However, a class can **explicitly** declare its intent using the `implements` keyword.

```ruby
class User implements Speaker, Serializable
  def speak
    "Hello"
  end
end
```

**Rules for `implements`:**
*   It is a **compile-time assertion**. The compiler will error if the class does not actually satisfy the interface.
*   It serves as documentation for developers.
*   It has **zero runtime overhead**.

### 9.4 The `any` Type

The empty interface `interface{}` is represented by the keyword `any`. 

```ruby
def log(thing : any)
  puts thing.to_s
end
```

### 9.5 Runtime Casting (`as` and `is_a?`)

Rugby provides safe mechanisms to work with interfaces at runtime.

*   **`obj.is_a?(Interface)`**: Returns `Bool`. Compiles to Go type assertion `_, ok := obj.(Interface)`.
*   **`obj.as?(Interface)`**: Returns an optional `Interface?`. Returns `nil` if the cast fails.
*   **`obj.as(Interface)`**: Force-casts to the interface. **Panics** if the cast fails.

### 9.6 Standard Interfaces

Rugby maps common Ruby concepts to idiomatic Go interfaces:

| Rugby Interface | Go Equivalent | Method |
|-----------------|---------------|--------|
| `Stringer`      | `fmt.Stringer`| `to_s` -> `String()` |
| `Error`         | `error`       | `message` -> `Error()` |
| `Closer`        | `io.Closer`   | `close` -> `Close()` |

---

## 10. Visibility & Naming

### 10.1 Core principle

**Inside a Rugby module, everything is usable.**

The `pub` keyword controls what becomes visible when the compiled Go package is imported by other Go (or Rugby) modules.

* `pub` = export to Go (uppercase in output)
* no `pub` = internal to package (lowercase in output)

There is no Ruby-style `private`/`public`.

### 10.2 What can be `pub`

**Functions:**

```ruby
def helper(x : Int) -> Int    # internal
  x * 2
end

pub def double(x : Int) -> Int  # exported
  helper(x)
end
```

**Classes:**

```ruby
pub class Counter
  pub def inc
    @n += 1
  end
end
```

* Class must be `pub` to be usable from Go
* Methods intended for Go must also be `pub`
* `pub def` inside non-`pub` class is a compile error

**Interfaces:**

```ruby
pub interface Greeter
  def greet(name : String) -> String
end
```

### 10.3 Rugby naming conventions

These are **style rules**, not visibility rules:

* Types (`class`, `interface`) → `CamelCase`
* Functions, methods, variables → `snake_case`
* Capitalization in Rugby **never** controls visibility

### 10.4 Go name generation

Rugby names are rewritten to idiomatic Go names:

| Rugby source | pub? | Go output |
|--------------|------|-----------|
| `def parse_json` | no | `parseJSON` |
| `pub def parse_json` | yes | `ParseJSON` |
| `def user_id` | no | `userID` |
| `pub def user_id` | yes | `UserID` |

### 10.5 Acronym list

To keep output Go-idiomatic, the compiler uses an acronym table.

**Standard acronyms:**

`id`, `url`, `uri`, `http`, `https`, `json`, `xml`, `api`, `uuid`, `ip`, `tcp`, `udp`, `sql`, `tls`, `ssh`, `cpu`, `gpu`

Rules:

* Exported: `id` → `ID`, `http` → `HTTP`
* Unexported: first-part acronym `http_*` → `http*`, later-part `*_id` → `*ID`

These mappings are currently standardized in the compiler.

### 10.6 Name collision rules

Collisions occur when different Rugby identifiers normalize to the same Go identifier:

* `foo_bar` and `foo__bar`

**Rule:** Collisions are compile-time errors showing:

* Both Rugby names and locations
* The resulting Go name
* Suggested fix

### 10.7 Reserved words

If a Go identifier would be a keyword (`type`, `var`, `func`) or conflict with generated names (`main`, `init`, `NewTypeName`):

* **`pub`**: compile error (force rename)
* **internal**: escape with trailing `_` (e.g., `type` → `type_`)

---

## 11. Go Interop

### 11.1 Imports

```ruby
import net/http
import encoding/json as json
```

* `import a/b` → Go `import "a/b"`
* `import a/b as x` → Go `import x "a/b"`

### 11.2 Calling Go functions

Rugby calls Go packages with dot syntax:

```ruby
http.Get(url)
json.Marshal(data)
```

Snake_case maps to CamelCase for Go interop:

```ruby
io.read_all(r)  # compiles to io.ReadAll(r)
```

Rules:

* Only for imported Go packages/types
* Mapping is compile-time only
* Compiler error if ambiguous

### 11.3 Struct fields and methods

* `resp.Body` allowed as-is
* Optional: `resp.body` → `resp.Body` for Go types

### 11.4 Defer

```ruby
defer resp.Body.Close()
```

Compiles to:

```go
defer resp.Body.Close()
```

Rule: `defer` requires a call expression with parentheses.

---

## 12. Runtime Package

Rugby provides a Go runtime package (`rugby/runtime`) that gives Ruby-like ergonomics to Go's built-in types. This is what makes Rugby feel like Ruby while compiling to idiomatic Go.

### 12.1 Design principles

* **Wrap, don't reinvent:** Functions wrap Go stdlib, adding Ruby-style APIs
* **Type-safe generics:** Use Go 1.18+ generics for collections
* **Zero-cost when unused:** Only imported when Rugby code uses these methods
* **Predictable mapping:** Each Rugby method maps to a clear runtime function

### 12.2 Package organization

```
runtime/
├── array.go      # Array/slice methods
├── map.go        # Map methods
├── string.go     # String methods
├── int.go        # Integer methods
├── float.go      # Float methods
├── bytes.go      # Byte slice methods
└── conv.go       # Type conversions
```

### 12.3 Array methods (`Array[T]`)

**Iteration:**
* `each { |x| }` → `runtime.Each(arr, fn)`
* `each_with_index { |x, i| }` → `runtime.EachWithIndex(arr, fn)`

**Transformation:**
* `map { |x| }` → `runtime.Map(arr, fn)` → `[]R`
* `select { |x| }` / `filter { |x| }` → `runtime.Select(arr, fn)` → `[]T`
* `reject { |x| }` → `runtime.Reject(arr, fn)` → `[]T`
* `compact` → `runtime.Compact(arr)` → filters nil/zero values

**Search:**
* `find { |x| }` / `detect { |x| }` → `runtime.Find(arr, fn)` → `T?`
* `any? { |x| }` → `runtime.Any(arr, fn)` → `Bool`
* `all? { |x| }` → `runtime.All(arr, fn)` → `Bool`
* `none? { |x| }` → `runtime.None(arr, fn)` → `Bool`
* `include?(val)` / `contains?(val)` → `runtime.Contains(arr, val)` → `Bool`

**Aggregation:**
* `reduce(init) { |acc, x| }` → `runtime.Reduce(arr, init, fn)` → `R`
* `sum` → `runtime.Sum(arr)` (numeric arrays)
* `min` / `max` → `runtime.Min(arr)` / `runtime.Max(arr)` → `T?`

**Access:**
* `first` → `runtime.First(arr)` → `T?`
* `last` → `runtime.Last(arr)` → `T?`
* `length` / `size` → `len(arr)` (inlined)
* `empty?` → `len(arr) == 0` (inlined)

**Mutation:**
* `reverse!` → `runtime.Reverse(arr)` (in-place)
* `sort!` → `runtime.Sort(arr)` (in-place, requires `Ordered`)
* `sort_by! { |x| }` → `runtime.SortBy(arr, fn)`

**Non-mutating variants:**
* `reverse` → `runtime.Reversed(arr)` → new slice
* `sort` → `runtime.Sorted(arr)` → new slice

### 12.4 Map methods (`Map[K, V]`)

**Iteration:**
* `each { |k, v| }` → `runtime.MapEach(m, fn)`
* `each_key { |k| }` → `runtime.MapEachKey(m, fn)`
* `each_value { |v| }` → `runtime.MapEachValue(m, fn)`

**Access:**
* `keys` → `runtime.Keys(m)` → `[]K`
* `values` → `runtime.Values(m)` → `[]V`
* `length` / `size` → `len(m)` (inlined)
* `empty?` → `len(m) == 0` (inlined)
* `has_key?(k)` / `key?(k)` → `_, ok := m[k]` (inlined)
* `fetch(k, default)` → `runtime.Fetch(m, k, default)` → `V`

**Transformation:**
* `select { |k, v| }` → `runtime.MapSelect(m, fn)` → `Map[K, V]`
* `reject { |k, v| }` → `runtime.MapReject(m, fn)` → `Map[K, V]`
* `merge(other)` → `runtime.Merge(m, other)` → `Map[K, V]`

### 12.5 String methods

**Query:**
* `length` / `size` → `len(s)` (inlined, byte length)
* `char_length` → `runtime.CharLength(s)` (rune count)
* `empty?` → `s == ""` (inlined)
* `include?(sub)` / `contains?(sub)` → `strings.Contains` (inlined)
* `start_with?(prefix)` → `strings.HasPrefix` (inlined)
* `end_with?(suffix)` → `strings.HasSuffix` (inlined)

**Transformation:**
* `upcase` → `strings.ToUpper` (inlined)
* `downcase` → `strings.ToLower` (inlined)
* `strip` → `strings.TrimSpace` (inlined)
* `lstrip` / `rstrip` → `strings.TrimLeft` / `TrimRight`
* `replace(old, new)` → `strings.ReplaceAll` (inlined)
* `reverse` → `runtime.StringReverse(s)`

**Splitting/Joining:**
* `split(sep)` → `strings.Split` (inlined)
* `chars` → `runtime.Chars(s)` → `[]String` (splits into characters)
* `lines` → `strings.Split(s, "\n")` (inlined)
* `bytes` → `[]byte(s)` (inlined)

**Conversion:**
* `to_i` → `runtime.MustAtoi(s)` (panics on failure)
* `to_i?` → `runtime.StringToInt(s)` → `Int?`
* `to_f?` → `runtime.StringToFloat(s)` → `Float?`

### 12.6 Integer methods

**Predicates:**
* `even?` → `n % 2 == 0` (inlined)
* `odd?` → `n % 2 != 0` (inlined)
* `zero?` → `n == 0` (inlined)
* `positive?` → `n > 0` (inlined)
* `negative?` → `n < 0` (inlined)

**Math:**
* `abs` → `runtime.Abs(n)` (or `math.Abs` for floats)
* `clamp(min, max)` → `runtime.Clamp(n, min, max)`

**Iteration:**
* `times { |i| }` → `runtime.Times(n, fn)`
* `upto(max) { |i| }` → `runtime.Upto(n, max, fn)`
* `downto(min) { |i| }` → `runtime.Downto(n, min, fn)`

**Conversion:**
* `to_s` → `strconv.Itoa(n)` (inlined)
* `to_f` → `float64(n)` (inlined)

### 12.7 Float methods

**Rounding:**
* `floor` → `math.Floor` (inlined)
* `ceil` → `math.Ceil` (inlined)
* `round` → `math.Round` (inlined)
* `truncate` → `math.Trunc` (inlined)

**Predicates:**
* `zero?` → `f == 0.0` (inlined)
* `positive?` / `negative?` → comparison (inlined)
* `nan?` → `math.IsNaN(f)` (inlined)
* `infinite?` → `math.IsInf(f, 0)` (inlined)

**Conversion:**
* `to_i` → `int(f)` (inlined)
* `to_s` → `strconv.FormatFloat` (inlined)

### 12.8 Codegen integration

The compiler recognizes Rugby method calls and either:
1. **Inlines** simple operations (predicates, length, type casts)
2. **Emits runtime calls** for complex operations

Example:
```ruby
nums = [1, 2, 3, 4, 5]
evens = nums.select { |n| n.even? }
sum = evens.reduce(0) { |acc, n| acc + n }
```

Compiles to:
```go
nums := []int{1, 2, 3, 4, 5}
evens := runtime.Select(nums, func(n int) bool { return n % 2 == 0 })
sum := runtime.Reduce(evens, 0, func(acc, n int) int { return acc + n })
```

### 12.9 Global functions (kernel)

Rugby provides top-level functions similar to Ruby's Kernel methods. These are available without qualification and compile to `runtime.*` calls.

**I/O:**
* `puts(args...)` → `runtime.Puts(args...)` - print with newline
* `print(args...)` → `runtime.Print(args...)` - print without newline
* `p(args...)` → `runtime.P(args...)` - debug print with inspect

**Input:**
* `gets` → `runtime.Gets()` - read line from stdin

**Program control:**
* `exit(code)` → `runtime.Exit(code)` - exit with status code
* `exit` → `runtime.Exit(0)` - exit successfully
* `sleep(seconds)` → `runtime.Sleep(seconds)` - pause execution

**Utilities:**
* `rand(n)` → `runtime.RandInt(n)` - random int [0, n)
* `rand` → `runtime.RandFloat()` - random float [0.0, 1.0)

### 12.10 Import generation

The compiler automatically adds `import "rugby/runtime"` when any runtime functions are used. All kernel functions and stdlib methods go through the runtime package for consistency.

## 13. Concurrency

Rugby treats Go's concurrency primitives as first-class citizens, providing Ruby-like syntax for high-performance concurrent programming.

### 13.1 Goroutines (`go`)

The `go` keyword executes a call in a new goroutine.

```ruby
go fetch_url(url)
go do
  puts "running in background"
end
```

**Compilation:**
*   `go func_call()` → `go funcCall()`
*   `go do ... end` → `go func() { ... }() (anonymous goroutine)`

### 13.2 Channels (`Chan[T]`)

Channels are typed pipes used for communication and synchronization.

**Syntax:**
*   `ch = Chan[Int].new(buffer_size)` (Buffered)
*   `ch = Chan[Int].new` (Unbuffered)
*   `ch << val` (Send)
*   `val = ch.receive` (Receive)
*   `val, ok = ch.receive?` (Safe receive)

**Compilation:**
*   `make(chan T, n)`
*   `ch <- val`
*   `val := <-ch`

### 13.3 Select

The `select` statement allows a goroutine to wait on multiple communication operations.

```ruby
select
when val = ch1.receive
  puts "received #{val}"
when ch2 << 42
  puts "sent to ch2"
else
  puts "no communication"
end
```

**Compilation:**
Maps directly to Go's `select` block.

---

## 14. Diagnostics

* Type errors and unresolved identifiers are compile-time errors
* Interop mapping errors (e.g., `io.read_all` not found) must show:
  * Rugby source span
  * Intended Go package/type
  * Suggested candidates (`ReadAll`)


