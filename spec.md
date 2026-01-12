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
| `Chan[T]`    | `chan T`    |
| `Task[T]`    | `struct` (internal) |

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

### 4.2.2 Channel Type

`Chan[T]` is a typed channel for concurrent communication between goroutines.

**Creation:**
```ruby
ch = Chan[Int].new(10)  # buffered channel with capacity 10
ch = Chan[Int].new      # unbuffered channel
```

**Compilation:** `Chan[T]` compiles to Go's `chan T`.

See Section 13.2 for full channel operations.

### 4.2.3 Task Type

`Task[T]` represents a concurrent computation that will produce a value of type `T`. Tasks are created with `spawn` and consumed with `await`.

```ruby
t = spawn { expensive_work() }  # t : Task[Result]
result = await t                 # result : Result
```

**Compilation:** `Task[T]` compiles to an internal struct containing a result channel.

See Section 13.4 for spawning and awaiting tasks.

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

Rugby requires explicit handling of optionals. "Truthiness" is **not** supported (see 5.2). Instead, use operators and patterns below.

#### 4.4.1 Optional Operators

Rugby provides two operators for working with optionals:

| Operator | Name | Description |
|----------|------|-------------|
| `??` | Nil coalescing | Unwrap or use default: `T? ?? T → T` |
| `&.` | Safe navigation | Call method if present: `T?&.method → R?` |

**Examples:**

```ruby
# Nil coalescing - unwrap or default
name = find_user(id)&.name ?? "Anonymous"
port = config.get("port") ?? 8080
user = find_user(id) ?? guest_user

# Safe navigation - chain through optionals
city = user&.address&.city ?? "Unknown"
len = input&.strip&.length ?? 0

# Combined
title = article&.author&.name ?? "Unknown Author"
```

**Lowering:**

```go
// name = find_user(id)&.name ?? "Anonymous"
tmp := findUser(id)
var name string
if tmp.Ok {
    name = tmp.Value.Name
} else {
    name = "Anonymous"
}
```

**Rules:**
* `??` requires left side to be `T?` and right side to be `T`
* `&.` requires left side to be `T?`, returns `R?`
* Using `??` or `&.` on non-optional types is a compile error

#### 4.4.2 Optional Methods

For more complex operations, optionals provide methods:

| Method | Returns | Description |
|--------|---------|-------------|
| `ok?` / `present?` | `Bool` | Is value present? |
| `nil?` / `absent?` | `Bool` | Is value absent? |
| `unwrap` | `T` | Unwrap, panic if absent |
| `map { \|x\| expr }` | `R?` | Transform if present |
| `flat_map { \|x\| expr }` | `R?` | Chain optionals (expr returns `R?`) |
| `each { \|x\| ... }` | `nil` | Execute block if present |

**Examples:**

```ruby
# Transform value if present
upper = name&.upcase  # simpler with &.
upper = name.map { |n| n.upcase }  # equivalent with method

# Chain optionals
city = user.flat_map { |u| u.address }.flat_map { |a| a.city }

# Side effect if present
user.each { |u| log.info("Found: #{u.name}") }

# Unwrap or panic (use sparingly)
value = required_field.unwrap
```

#### 4.4.3 The `if let` Pattern

For conditional unwrapping with a scoped binding, use `if let`:

```ruby
if let user = find_user(id)
  puts user.name
end
```

Compiles to `if user, ok := findUser(id); ok { ... }`

This is the preferred pattern when you need to use the unwrapped value conditionally.

#### 4.4.4 Tuple Unpacking

Optionals can be unpacked into a value and boolean using tuple assignment:

```ruby
user, ok = find_user(id)
if ok
  puts user.name
end
```

**Semantics:**
- `find_user(id)` returns `User?`
- Single assignment: `user = find_user(id)` → `user : User?`
- Tuple assignment: `user, ok = find_user(id)` → `user : User, ok : Bool`

When unpacked, `user` is the unwrapped type (not optional). If `ok` is false, `user` holds the zero value.

**Lowering:**
```go
user, ok := findUser(id)
if ok {
    fmt.Println(user.Name)
}
```

This mirrors Go's comma-ok idiom and avoids needing `.unwrap` after checking.

#### 4.4.5 Choosing a Pattern

| Pattern | Best for |
|---------|----------|
| `expr ?? default` | Providing a fallback value |
| `expr&.method` | Safe navigation through optionals |
| `if let u = expr` | Conditional use with scoped binding |
| `val, ok = expr` | When you need the boolean separately |
| `expr.unwrap` | When absence is a bug (will panic) |
| `expr.map { }` | Complex transformations |

### 4.5 The `nil` Keyword

Rugby has a `nil` keyword representing absence. Unlike Ruby where any value can be `nil`, Rugby restricts where `nil` is valid.

#### 4.5.1 Where `nil` is Valid

| Type | Can be `nil`? | Notes |
|------|---------------|-------|
| Value types (`Int`, `String`, `Bool`, `Float`) | No | Use `T?` for optional values |
| Reference types (`User`, `Writer`) | No | Use `T?` for optional references |
| Optional types (`T?`) | Yes | `nil` means absent |
| `error` | Yes | `nil` means no error (success) |

#### 4.5.2 Examples

**Optional types:**
```ruby
def find_user(id : Int) -> User?
  if id < 0
    return nil  # OK: User? can be nil
  end
  User.new("Alice")
end

user : User? = nil  # OK
name : String? = nil  # OK
```

**Non-optional types (compile errors):**
```ruby
def get_user -> User
  nil  # ERROR: User cannot be nil, use User?
end

x : Int = nil  # ERROR: Int cannot be nil, use Int?
```

**Error type (special case):**
```ruby
def save(data : String) -> error
  if data.empty?
    return fmt.errorf("empty data")
  end
  nil  # OK: nil means success
end
```

#### 4.5.3 Checking for `nil`

For optional types, use tuple unpacking or `if let`:

```ruby
# Tuple unpacking (preferred)
user, ok = find_user(5)
if ok
  puts user.name
else
  puts "not found"
end

# Or if let
if let user = find_user(5)
  puts user.name
else
  puts "not found"
end

# Or check nil directly
result = find_user(5)
if result.nil?
  puts "not found"
end
```

For the `error` type, use direct comparison:

```ruby
err = save("data")
if err != nil
  puts "failed: #{err}"
end

if err == nil
  puts "success"
end
```

#### 4.5.4 Why This Design?

1. **Safety:** Forcing `T?` for optional values makes nullability explicit in the type system.
2. **Go compatibility:** Maps cleanly to Go's zero values and pointer semantics.
3. **Ruby familiarity:** `nil` keyword is familiar, but scoped to where it makes sense.
4. **Error convention:** `error` being nullable matches Go's idiomatic error handling.

**Lowering:**

* `T?` with `nil` → `Option[T]{Ok: false}` or `(T{}, false)`
* `error` with `nil` → Go `nil` directly

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

Use `if let` to handle optionals and type assertions safely.

```ruby
if let user = find_user(id)
  puts user.name
end
```

**Type assertion with `as`:**

```ruby
if let w = obj.as(Writer)
  w.write("hello")
end

# Or with default
w = obj.as(Writer) ?? fallback_writer

# Or unwrap (panics if wrong type)
w = obj.as(Writer).unwrap
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

Rugby supports Ruby-style "command syntax" where parentheses are optional for function and method calls.

**With parentheses (always valid):**

```ruby
inc()
add(1, 2)
puts("hello")
user.save()
```

**Without parentheses (command syntax):**

```ruby
puts "hello"       # puts("hello")
add 1, 2           # add(1, 2)
foo.bar baz        # foo.bar(baz)
print x + 1        # print(x + 1)
```

**Rules:**

1. Arguments are comma-separated expressions
2. The argument list ends at: newline, `do`, `{`, or logical operators (`and`, `or`)
3. Arithmetic and comparison operators are included in arguments:
   - `puts x + 1` → `puts(x + 1)`
   - `puts x == y` → `puts(x == y)`
4. Logical operators terminate arguments:
   - `puts x or y` → `puts(x) or y`
5. Unary minus disambiguation uses whitespace:
   - `foo -1` → `foo(-1)` (space before `-`, no space after)
   - `foo - 1` → `(foo - 1)` (space on both sides = binary)
6. Collection literals with space are arguments:
   - `foo [1, 2]` → `foo([1, 2])` (array argument)
   - `foo[1]` → index expression (no space = indexing)
7. **Properties:** `user.age` (getter) and `user.age = 5` (setter) do not use parentheses
8. For clarity in complex expressions, parentheses are recommended

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

### 7.6 Inheritance & Specialization

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

### 7.7 Explicit Implementation (Optional)

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

### 7.8 Polymorphism & Interfaces

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
*   **`obj.as(Interface)`**: Returns `Interface?` (optional). Use with `if let`, `??`, or `.unwrap`.

```ruby
# Check type (predicate)
if obj.is_a?(Writer)
  puts "it's a writer"
end

# Cast with if let (preferred)
if let w = obj.as(Writer)
  w.write("hello")
end

# Cast with fallback
w = obj.as(Writer) ?? default_writer

# Cast with unwrap (panics if wrong type)
w = obj.as(Writer).unwrap
```

**Lowering:**

```go
// obj.as(Writer) compiles to:
func asWriter(obj any) (Writer, bool) {
    w, ok := obj.(Writer)
    return w, ok
}
```

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
* Predicate methods → `snake_case?` (must return `Bool`)

**The `?` suffix rule:**

Method names ending in `?` **must** return `Bool`. This is strictly enforced:

```ruby
def empty? -> Bool    # OK: predicate
def valid? -> Bool    # OK: predicate
def even? -> Bool     # OK: predicate
```

For operations that might fail, use error returns instead:

```ruby
def to_i -> (Int, error)     # fallible conversion
def parse -> (Data, error)   # fallible parsing
```

This keeps `?` unambiguous: when you see `method?`, you know it returns a boolean.

**Other conventions:**

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
├── conv.go       # Type conversions
├── task.go       # Task[T] and spawn/await
├── scope.go      # Structured concurrency scope
└── channel.go    # Channel helpers (try_receive)
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

**Mutation (in-place):**
* `reverse` → `runtime.Reverse(arr)` (in-place)
* `sort` → `runtime.Sort(arr)` (in-place, requires `Ordered`)
* `sort_by { |x| }` → `runtime.SortBy(arr, fn)` (in-place)

**Non-mutating (returns copy):**
* `reversed` → `runtime.Reversed(arr)` → new slice
* `sorted` → `runtime.Sorted(arr)` → new slice

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
* `to_i` → `runtime.StringToInt(s)` → `(Int, error)`
* `to_f` → `runtime.StringToFloat(s)` → `(Float, error)`

Use with `!` or `rescue`:
```ruby
n = s.to_i()!           # propagate error
n = s.to_i() rescue 0   # default on failure
```

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

Rugby provides Ruby-like ergonomics for Go's concurrency model while preserving Go's semantics and performance characteristics.

Concurrency is built on two layers:

1. **Go primitives:** goroutines, channels, select
2. **Structured concurrency:** `spawn`, `await`, and `concurrently` scopes

The structured layer compiles to Go primitives and exists to prevent leaks, provide deterministic cleanup, and make cancellation and error propagation consistent.

---

### 13.1 Goroutines (`go`)

The `go` keyword executes a call in a new goroutine.

```ruby
go fetch_url(url)

go do
  puts "running in background"
end
```

**Compilation:**
* `go func_call()` → `go funcCall()`
* `go do ... end` → `go func() { ... }()`

**Notes:**
* `go` is low-level and does not provide a return value.
* For "run and await result" semantics, prefer `spawn` + `await` (see 13.4).

---

### 13.2 Channels (`Chan[T]`)

Channels are typed pipes for communication and synchronization.

**Creation:**
```ruby
ch = Chan[Int].new(10)   # buffered (capacity 10)
ch = Chan[Int].new       # unbuffered
```

**Operations:**
```ruby
ch << val            # send (blocks if full/unbuffered)
val = ch.receive     # receive (blocks; returns zero value if closed)
val = ch.try_receive # non-blocking receive, returns T?
ch.close             # close the channel
```

**`try_receive` returns `T?`** for consistency with Rugby's optional pattern:
```ruby
if let msg = ch.try_receive
  puts "got: #{msg}"
else
  puts "channel empty or closed"
end

# Or with default
msg = ch.try_receive ?? "default"
```

**Compilation:**
| Rugby | Go |
|-------|-----|
| `Chan[T].new(n)` | `make(chan T, n)` |
| `Chan[T].new` | `make(chan T)` |
| `ch << val` | `ch <- val` |
| `ch.receive` | `<-ch` |
| `ch.try_receive` | wrapped `select` with default (returns `Option[T]`) |
| `ch.close` | `close(ch)` |

**Channel iteration:**

Channels can be iterated until closed:

```ruby
for msg in ch
  puts msg
end
```

Compiles to:
```go
for msg := range ch {
    runtime.Puts(msg)
}
```

---

### 13.3 Select

The `select` statement waits on multiple channel operations.

```ruby
select
when val = ch1.receive
  puts "received #{val}"
when ch2 << 42
  puts "sent to ch2"
else
  puts "no communication ready"
end
```

**Semantics:**
* Executes the first ready case.
* If multiple cases are ready, one is chosen at random.
* The `else` clause makes select non-blocking.
* Variables bound in `when` clauses (like `val`) are scoped to that branch.

**Note:** `when` in `select` binds variables from channel operations. This differs from `when` in `case` expressions (Section 5.2), which matches values. The keyword is reused for familiarity, but the semantics are distinct.

**Compilation:**
Maps directly to Go's `select { case ... }` block.

---

### 13.4 Tasks (`spawn` and `await`)

Tasks provide value-oriented concurrency: run work concurrently and retrieve its result.

#### 13.4.1 Spawning Tasks

```ruby
t = spawn do
  heavy_computation()
end
```

* `spawn` starts execution immediately in a goroutine.
* Returns a typed handle: `Task[T]` where `T` is the block's return type.
* Type inference works naturally:

```ruby
t = spawn { 42 }           # t : Task[Int]
t = spawn { "hello" }      # t : Task[String]
t = spawn { fetch_data() } # t : Task[Data] if fetch_data returns Data
```

**Compilation:**

`spawn` lowers to:
* Create a result channel
* Start a goroutine that evaluates the block and sends the result
* Return a `Task[T]` wrapping the channel

#### 13.4.2 Awaiting Tasks

```ruby
value = await t
```

**Syntax:**
* `await expr` — awaits a task expression
* `await(expr)` — equivalent; parentheses are optional but required when chaining with `!`

`await` is a keyword that binds to the immediately following expression:
* `await t` — await task `t`
* `await get_task()` — await the result of `get_task()`
* `(await t).name` — await `t`, then access `.name` on the result
* `await(t)!` — await and propagate error (parentheses required for `!`)

**Semantics:**
* `await` blocks until the task completes and returns its value.
* Awaiting a completed task returns immediately.
* Awaiting a task multiple times is a compile error (tasks complete once).

**Error-returning tasks:**

If the block returns `(T, error)`, awaiting returns `(T, error)`:

```ruby
t = spawn do
  os.read_file("data.txt")  # returns (Bytes, error)
end

data, err = await t
return nil, err if err != nil
```

**Integration with `!` operator:**

The `!` operator (Section 15.3) works with `await`:

```ruby
def load_all -> (Array[Bytes], error)
  t1 = spawn { os.read_file("a.txt") }
  t2 = spawn { os.read_file("b.txt") }

  a = await(t1)!  # propagate error if t1 failed
  b = await(t2)!  # propagate error if t2 failed

  [a, b], nil
end
```

#### 13.4.3 Panics in Tasks

Panics inside a spawned task behave like Go panics:
* Panics are not captured by default.
* A panic in any goroutine terminates the program unless recovered.

---

### 13.5 Structured Concurrency (`concurrently`)

Unscoped concurrency leads to goroutine leaks, unclear cancellation, and inconsistent error handling. Rugby provides `concurrently` for deterministic, scoped concurrency.

A `concurrently` block owns all tasks spawned within it and ensures cleanup when the block exits.

#### 13.5.1 Basic Usage

```ruby
concurrently do |scope|
  a = scope.spawn { fetch_a() }
  b = scope.spawn { fetch_b() }

  x = await a
  y = await b

  puts x + y
end
```

**Rules:**
1. `concurrently` introduces a scope object.
2. `scope.spawn { ... }` creates a task owned by the scope.
3. On block exit (normal, `return`, or error), the scope:
   * Cancels its context (signals tasks to stop)
   * Waits for all spawned tasks to complete
4. Tasks can be awaited in any order.

This makes task lifetime explicit: tasks cannot outlive their scope.

**Return value:**

`concurrently` is an expression. Its value is the last expression in the block:

```ruby
# Returns a value
total = concurrently do |scope|
  a = scope.spawn { fetch_a() }
  b = scope.spawn { fetch_b() }
  (await a) + (await b)  # block evaluates to this sum
end

# Used as statement (no return value needed)
concurrently do |scope|
  scope.spawn { log_metrics() }
  scope.spawn { send_notifications() }
end
```

If tasks return errors and you use `!`, the block returns `(T, error)`:

```ruby
result, err = concurrently do |scope|
  data = await(scope.spawn { fetch_data() })!
  process(data), nil
end
```

#### 13.5.2 Cancellation Context

Each scope provides a cancellation context via `scope.ctx`:

```ruby
concurrently do |scope|
  t = scope.spawn do
    long_running_work(scope.ctx)
  end

  # If we return early, scope.ctx is cancelled
  return if should_abort?

  await t
end
```

**Semantics:**
* When the scope exits, its context is cancelled.
* Tasks should check the context cooperatively (Go-style).
* Rugby does not forcibly terminate goroutines.

**Compilation:**

`concurrently` lowers to:
```go
ctx, cancel := context.WithCancel(parentCtx)
defer cancel()
var wg sync.WaitGroup
// ... spawned tasks use ctx and increment wg
wg.Wait()
```

#### 13.5.3 Error Propagation

If any scoped task returns an error:
1. The scope cancels its context (requesting siblings stop).
2. The scope waits for all tasks to finish.
3. The first error is returned from the `concurrently` block.

```ruby
result, err = concurrently do |scope|
  a = scope.spawn { fetch_a() }  # -> (A, error)
  b = scope.spawn { fetch_b() }  # -> (B, error)

  val_a = await(a)!  # propagate error
  val_b = await(b)!  # propagate error

  process(val_a, val_b)
end
```

Using `!` inside `concurrently` triggers early exit from the block, cancelling sibling tasks.

#### 13.5.4 Detached Tasks

Tasks spawned via `spawn` (not `scope.spawn`) are **detached** and may outlive the current function:

```ruby
spawn do
  background_worker()  # runs independently
end
```

**Guideline:** Prefer `concurrently` for tasks that must complete safely. Use detached `spawn` sparingly for fire-and-forget work.

#### 13.5.5 Escape Restriction

A task handle from `scope.spawn` must not escape the `concurrently` block:

```ruby
var leaked : Task[Int]

concurrently do |scope|
  leaked = scope.spawn { 42 }  # ERROR: task escapes scope
end

await leaked  # would await after scope cleanup
```

This is a compile-time error. It prevents awaiting tasks after their scope has ended.

---

### 13.6 Choosing a Pattern

| Need | Use |
|------|-----|
| Fire-and-forget background work | `go do ... end` |
| Low-level channel communication | `Chan[T]`, `select` |
| Run work and get result | `spawn` + `await` |
| Multiple concurrent tasks with cleanup | `concurrently` |
| Cancellable, scoped work | `concurrently` with `scope.ctx` |

**Examples:**

**Simple parallel fetch:**
```ruby
def fetch_both(url1 : String, url2 : String) -> (String, String, error)
  concurrently do |scope|
    t1 = scope.spawn { http.get(url1) }
    t2 = scope.spawn { http.get(url2) }

    r1 = await(t1)!
    r2 = await(t2)!

    r1.body, r2.body, nil
  end
end
```

**Worker pool with channels:**
```ruby
def process_items(items : Array[Item])
  work = Chan[Item].new(100)
  done = Chan[Bool].new

  # Start workers
  4.times do
    go do
      for item in work
        process(item)
      end
      done << true
    end
  end

  # Feed work
  items.each { |item| work << item }
  work.close

  # Wait for workers
  4.times { done.receive }
end
```

**Timeout pattern:**
```ruby
select
when result = work_chan.receive
  puts "got result: #{result}"
when timeout_chan.receive
  puts "timed out"
end
```

---

### 13.7 Compilation Summary

| Rugby | Go |
|-------|-----|
| `go expr` | `go expr` |
| `go do ... end` | `go func() { ... }()` |
| `Chan[T].new(n)` | `make(chan T, n)` |
| `spawn { expr }` | goroutine + channel for result |
| `await t` | receive from task's channel |
| `concurrently do \|s\| ... end` | `context.WithCancel` + `sync.WaitGroup` |
| `scope.spawn { ... }` | goroutine tracked by WaitGroup |
| `scope.ctx` | `context.Context` |

---

## 14. Diagnostics

* Type errors and unresolved identifiers are compile-time errors
* Interop mapping errors (e.g., `io.read_all` not found) must show:
  * Rugby source span
  * Intended Go package/type
  * Suggested candidates (`ReadAll`)

---

## 15. Errors

Rugby follows Go's philosophy: **errors are values**, returned explicitly from functions and handled locally. Rugby does **not** use exceptions for ordinary failures.

Rugby provides syntax sugar to keep error handling concise while compiling to idiomatic Go.

---

### 15.1 The `error` Type

Rugby includes a built-in `error` type that maps directly to Go's `error` interface.

**Rules:**
* A function may return `error` as a sole return value, or as part of a multi-return tuple.
* `nil` represents "no error."

```ruby
def write_file(path : String, data : Bytes) -> error
  os.write_file(path, data)
end

def read_file(path : String) -> (Bytes, error)
  os.read_file(path)
end
```

---

### 15.2 Fallible Function Signatures

The standard pattern for fallible operations:

* `-> (T, error)` — functions that produce a value or an error
* `-> error` — procedures that may fail

```ruby
def parse_int(s : String) -> (Int, error)
  # ...
end

def close -> error
  # ...
end
```

---

### 15.3 Postfix Bang Operator (`!`)

Rugby provides a postfix `!` operator for concise error propagation, similar to Rust's `?` operator. It unwraps the result or propagates the error to the caller.

#### 15.3.1 Syntax

```ruby
call_expr!
```

Examples:

```ruby
data = os.read_file(path)!
os.mkdir_all(dir)!
```

#### 15.3.2 Valid Operands

Postfix `!` is valid on expressions whose return type includes a trailing `error`:

* `error`
* `(T, error)`
* `(A, B, ..., error)`

**Accepted expression forms:**

* **Call expressions**: `call()!`, `obj.method(arg)!`
* **Selector expressions**: `obj.method!` (equivalent to `obj.method()!`)
* **Identifiers**: `func_name!` (equivalent to `func_name()!`)
* **`await` expressions**: `await(expr)!` when the task returns `(T, error)`

The latter two are syntactic sugar that convert to zero-argument calls.

```ruby
data = os.read_file(path)!  # explicit call
data = resp.json!           # shorthand for resp.json()!
data = await(task)!         # await with error propagation
```

**Compile-time errors:**
* Using `!` on a non-callable expression (e.g., `42!` or `"string"!`)
* Using `!` on a call that does not return `error`
* Using `!` on `await expr` without parentheses (use `await(expr)!`)

#### 15.3.3 Enclosing Function Requirement

Using `!` requires the enclosing function to return `error`:

* `-> error`
* `-> (T, error)`
* `-> (A, B, ..., error)`

If the enclosing function does not return `error`, using `!` is a compile-time error.

**Exception:** In `main` or top-level script context, `!` has special behavior (see 15.6).

#### 15.3.4 Semantics

* If the call succeeds (`err == nil`), execution continues.
* If the call fails (`err != nil`), the current function returns immediately with that error.

**Unwrapping:**
* `(T, error)` → `call!` evaluates to the unwrapped `T`
* `error` → `call!` is used only for control-flow effect

#### 15.3.5 Precedence

Postfix `!` binds tighter than all operators except member access and calls. It applies to the completed call expression:

* `foo(bar)!` — call, then bang
* `a.b().c()!` — bang applies to `c()`
* `!foo()!` — prefix negation of unwrapped value (valid only if result is `Bool`)

#### 15.3.6 Chained Calls

Multiple `!` operators can appear in a chain:

```ruby
user = load_file(path)!.parse_json()!.get_user()!
```

Each `!` unwraps its call before the next selector executes.

**Lowering:**

The compiler generates sequential temporaries:

```go
tmp1, err := loadFile(path)
if err != nil {
    return User{}, err
}
tmp2, err := tmp1.parseJSON()
if err != nil {
    return User{}, err
}
user, err := tmp2.getUser()
if err != nil {
    return User{}, err
}
```

---

### 15.4 The `rescue` Keyword

Rugby provides `rescue` for inline error handling when you need more control than simple propagation.

#### 15.4.1 Inline Default Form

Provide a fallback value when a call fails:

```ruby
expr = fallible_call() rescue default_expr
```

If the call returns an error, `default_expr` is evaluated and used instead. Execution continues.

```ruby
port = env.get("PORT").to_i() rescue 8080
config = parse_json(text) rescue {}
data = os.read_file(path) rescue default_bytes
```

**Lowering:**
```go
port, err := strconv.Atoi(env.Get("PORT"))
if err != nil {
    port = 8080
}
```

**Type rule:** `default_expr` must match the success type `T`.

#### 15.4.2 Block Form

For multi-statement error handling, use `rescue do ... end`:

```ruby
expr = fallible_call() rescue do
  log.warn("operation failed")
  compute_fallback()
end
```

The block's last expression becomes the value. Alternatively, use `return` for early exit.

```ruby
data = os.read_file(path) rescue do
  log.warn("using default config")
  default_config
end
```

**Lowering:**
```go
data, err := os.ReadFile(path)
if err != nil {
    log.Warn("using default config")
    data = defaultConfig
}
```

#### 15.4.3 Block Form with Error Binding

Bind the error to a variable using `=> name`:

```ruby
expr = fallible_call() rescue => err do
  log.warn("failed: #{err}")
  return nil, fmt.errorf("context: %w", err)
end
```

The variable `err` is in scope within the block.

```ruby
data = os.read_file(path) rescue => err do
  log.warn("read failed: #{err}")
  return nil, fmt.errorf("load_config: %w", err)
end
```

**Lowering:**
```go
data, err := os.ReadFile(path)
if err != nil {
    log.Warn(fmt.Sprintf("read failed: %v", err))
    return nil, fmt.Errorf("load_config: %w", err)
}
```

#### 15.4.4 Rules

1. `rescue` applies to call expressions returning `error` or `(T, error)`.
2. In inline form, `default_expr` must match type `T`.
3. In block form, the last expression becomes the value (unless `return` is used).
4. `rescue` and `!` are mutually exclusive on the same call.
5. The `=> name` binding creates a new variable scoped to the block, shadowing any outer variable with the same name.

#### 15.4.5 `rescue` on `error`-only Returns

For calls returning only `error` (no value), `rescue` executes for side effects:

```ruby
os.mkdir_all(dir) rescue => err do
  log.warn("mkdir failed: #{err}")
  # no value needed
end
```

**Lowering:**
```go
if err := os.MkdirAll(dir, 0755); err != nil {
    log.Warn(fmt.Sprintf("mkdir failed: %v", err))
}
```

The inline form is not allowed for `error`-only calls (no value to replace):

```ruby
os.mkdir_all(dir) rescue nil   # ERROR: inline rescue requires (T, error) return
```

---

### 15.5 Lowering Rules

#### A) `call!` where call returns `(T, error)`

Rugby:
```ruby
x = f(a, b)!
```

Go:
```go
x, err := f(a, b)
if err != nil {
    return <zero-values>, err
}
```

#### B) `call!` where call returns `error`

Rugby:
```ruby
f(a, b)!
```

Go:
```go
if err := f(a, b); err != nil {
    return <zero-values>, err
}
```

#### C) Multiple return values

If the enclosing function returns multiple values, the compiler constructs appropriate zero values:

Rugby:
```ruby
def open_user(id : Int) -> (User, Bool, error)
  u = load_user(id)!
  u, true, nil
end
```

Go:
```go
func openUser(id int) (User, bool, error) {
    u, err := loadUser(id)
    if err != nil {
        return User{}, false, err
    }
    return u, true, nil
}
```

**Zero values:**
* Numeric: `0`
* Bool: `false`
* String: `""`
* Slices, maps, funcs, interfaces, pointers: `nil`
* Structs: `T{}`

---

### 15.6 Script Mode and `main`

In bare script mode (see 2.1) and inside `def main`, there is no enclosing function that returns `error`.

**Rule:** `call!` prints the error to stderr and exits with status code `1`.

```ruby
# script.rg
data = os.read_file("config.json")!
puts data.length
```

**Lowering:**
```go
func main() {
    data, err := os.ReadFile("config.json")
    if err != nil {
        runtime.Fatal(err)
    }
    runtime.Puts(len(data))
}
```

`runtime.Fatal(err)` prints a readable error message to stderr and calls `os.Exit(1)`.

---

### 15.7 Explicit Handling

For cases where `!` and `rescue` don't fit, handle errors explicitly:

```ruby
data, err = os.read_file(path)
if err != nil
  log.error("failed to read #{path}: #{err}")
  return nil, err
end
```

This is standard Go-style error handling and always works.

---

### 15.8 Panics (Programmer Errors)

For unrecoverable programmer errors or impossible states, use `panic`:

```ruby
panic "unreachable"
panic "invalid state: #{state}"
```

**Lowering:**
```go
panic("unreachable")
panic(fmt.Sprintf("invalid state: %v", state))
```

**Guideline:** Use `panic` only for invariants and bugs. Use returned `error` for expected failures.

---

### 15.9 Error Utilities

Rugby exposes common Go error utilities via the runtime:

**`error_is?(err, target)`**

Tests if an error matches a target (wrapping-aware):

```ruby
if error_is?(err, os.ErrNotExist)
  puts "file not found"
end
```

Compiles to `errors.Is(err, target)`.

**`error_as(err, Type)`**

Extracts a specific error type. Returns `Type?` (optional):

```ruby
if let path_err = error_as(err, os.PathError)
  puts "path error: #{path_err.path}"
end
```

Compiles to:
```go
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    // use pathErr
}
```

Note: `error_as` returns an optional (not an error), so use `if let` for unwrapping.

---

### 15.10 Common Patterns

**Retry with fallback:**
```ruby
def fetch_config -> (Config, error)
  data = http.get(primary_url) rescue => err do
    log.warn("primary failed: #{err}")
    http.get(backup_url)!
  end
  parse_config(data)!
end
```

**Wrap and propagate:**
```ruby
def load_user(id : Int) -> (User, error)
  row = db.query_row(sql, id) rescue => err do
    return nil, fmt.errorf("load_user(%d): %w", id, err)
  end
  parse_user(row)!
end
```

**Cleanup on error:**
```ruby
def process_file(path : String) -> error
  f = os.open(path)!
  defer f.close()

  process(f) rescue => err do
    log.error("processing failed: #{err}")
    return err
  end

  nil
end
```

**Ignore error, use default:**
```ruby
timeout = config.get("timeout").to_i() rescue 30
```

---

### 15.11 Determinism

All error handling is compile-time, explicit, and deterministic. There is no implicit exception control flow, stack unwinding, or reflection-based error handling.

