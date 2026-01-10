# Rugby Language Spec (MVP v0.1)

**Goal:** Ruby-ish surface syntax that compiles to idiomatic Go and makes using Go libraries feel native.

Rugby is **compiled**, **static**, and **predictable**.

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

  def inc!
    @n += 1
  end

  def value -> Int
    @n
  end
end

c = Counter.new(0)
c.inc!
c.inc!
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
"interpolation: #{name}"  # compiles to fmt.Sprintf or concat
```

String interpolation rules:
* Any expression is allowed inside `#{}`
* Non-string types use `fmt.Sprintf` with `%v` formatting
* Examples:
  * `"count: #{items.length}"` - method call
  * `"sum: #{a + b}"` - arithmetic
  * `"user: #{user.to_s}"` - explicit conversion

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
For the MVP, symbols compile to **Go strings**.
`:ok` → `"ok"`

*Future optimization:* Compile to integer constants for fast comparison.

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

### 4.3 Type annotations

Rugby follows Go's philosophy: rely on Go's type inference where possible.

**Variables:** Type annotations are optional. Without annotation, Go infers from the assignment:

```ruby
x = 5           # x := 5 (Go infers int)
y : Int64 = 5   # var y int64 = 5 (explicit type)
```

**Function parameters:** Type annotations are optional. Without annotation, parameters default to `interface{}`:

```ruby
# Untyped - uses interface{}
def add(a, b)
  a + b
end
# → func add(a interface{}, b interface{}) { ... }

# Typed parameters
def add(a : Int, b : Int) -> Int
  a + b
end
# → func add(a int, b int) int { return a + b }
```

**Instance variables:** Types are inferred from `initialize` parameter types:

```ruby
class User
  def initialize(name : String, age : Int)
    @name = name   # @name is string
    @age = age     # @age is int
  end
end
# → type User struct { name string; age int }
```

**Type checking:** Rugby does not perform type checking. Type mismatches are caught by the Go compiler.

### 4.4 Optionals (`T?`)

**Syntax:** `T?` is the universal syntax for "nullable" or "optional" values. The compiler abstracts the underlying Go representation to ensure idiomatic usage.

**Representation:**

1.  **Value Types** (`Int`, `Float`, `Bool`, `String`, Structs):
    *   **Return Values:** Compiles to `(T, bool)` (standard Go "comma-ok" idiom).
    *   **Storage (Fields, Arrays, Maps):** Compiles to `*T` (pointer to value) to represent nullability in data structures.
    *   **Local Variables:** Compiler determines representation (unpacked `val, ok` or `*T`).
    *   Example: `Int?` → `(int, bool)` (return) or `*int` (storage).

2.  **Reference Types** (`Array`, `Map`, Classes, Interfaces):
    *   Compiles to the underlying pointer/slice/map type.
    *   Uses Go's `nil` to represent absence.
    *   Example: `User?` → `*User` (where `User` compiles to `struct`)

**Usage:**

Conditionals unify these differences. `if x` checks for presence:

```ruby
# Works for both value types and reference types
def process(u : User?, score : Int?)
  if u        # Checks u != nil
    puts u.name
  end

  if score    # Checks the boolean flag
    puts score
  end
end
```

**Assignment with check:**

```ruby
if (n = s.to_i?)
  puts n
end
```
Compiles to `if n, ok := strconv.Atoi(s); ok { ... }`

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

```ruby
if cond
  ...
elsif other
  ...
else
  ...
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

**Case Expressions:**

Rugby supports `case` for pattern matching.

```ruby
case status
when 200
  puts "ok"
when 404
  puts "not found"
else
  puts "error"
end
```

**Compilation:**
Compiles to Go `switch` statement.
*   Values map to `case val:`
*   Multiple values `when 1, 2` map to `case 1, 2:`
*   Type matching (Future): `when String` maps to type switch.

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

Use blocks when you need to **transform data** or chain operations. Blocks in Rugby are strictly **anonymous functions** (lambdas).

**Syntax:**
* `do ... end`: Preferred for multi-line blocks.
* `{ ... }`: Preferred for single-line blocks.

**Semantics:**
* **Scope:** Creates a new local scope (Go function literal).
* **Return:** `return` exits the **block only** (returns a value to the iterator), unlike Ruby where it returns from the enclosing method.
* **Break/Next:** Supported via boolean return signals in the generated Go code.

```ruby
names = users.map do |u|
  break if u.is_admin? # Stops iteration
  next if u.inactive?  # Skips to next item
  u.name
end
```

**Compilation:**

1. **Iterative Blocks** (`each`, `times`, `upto`, `downto`):
   The callback returns `bool`. `false` signals `break`, `true` signals `next` (continue).
   ```go
   runtime.Each(items, func(item interface{}) bool {
       if condition { return false } // break
       return true // next
   })
   ```

2. **Transformation Blocks** (`map`, `select`, `reject`, `find`, `reduce`):
   The callback returns `(T, bool)` where the second value signals continuation.
   ```go
   runtime.Map(items, func(item interface{}) (interface{}, bool) {
       if condition { return nil, false } // break
       return result, true // next
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

Optional sugar (`try`/`?`) may come later, but must lower to explicit error checks.

---

## 7. Classes

### 7.1 Definition

```ruby
class User
  def initialize(name : String, age : Int)
    @name = name
    @age  = age
  end

  def greet -> String
    "hi #{@name}"
  end

  def birthday!
    @age += 1
  end
end
```

### 7.2 Go lowering

* `class User` → `type User struct { name string; age int }`
* `@name`/`@age` become struct fields (unexported by default)
* If `pub`, fields are exported

### 7.2.1 Instance variable type inference

Instance variable types are inferred by scanning **all methods** in the class.

```ruby
class User
  def initialize(name : String)
    @name = name
  end
  
  def set_age(age : Int)
    @age = age  # @age inferred as Int, added to struct
  end
end
```

Rules:
* All assignments to the same `@var` must have consistent types.
* Fields implied by non-`initialize` methods are created with Go's zero value.
* It is good practice, but not strictly required, to initialize all fields in `initialize`.

### 7.3 Constructors

* `initialize` generates a constructor function
* `User.new(...)` rewrites to constructor call
* If class is `pub`: `NewUser(...) *User`
* If internal: `newUser(...) *User`

**Note:** Constructors always return pointers (`*T`) for consistency. This allows mutation methods (`!`) to work without reassignment.

### 7.4 Methods and receivers

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

### 7.5 Methods with `!`

In Rugby, `!` is a **naming convention** only. It does not change compilation semantics (since all methods are already pointer receivers).

```ruby
def inc!
  @n += 1
end
```

* `!` is stripped from the generated Go function name to match Go idioms.
* `def inc!` → `func (c *Counter) inc()`
* `pub def inc!` → `func (c *Counter) Inc()`

This allows Rugby code to communicate "danger/mutation" (`save!`) while generating standard Go names.

### 7.6 Embedding (composition)

```ruby
class Service < Logger, Authenticator
  def run
    info("running")
    authenticate!
  end
end
```

Compiles to:

```go
type Service struct {
    Logger
    Authenticator
}
```

* `<` means embedding (Go struct embedding).
* Multiple embedding is supported (comma-separated).
* Name collisions in embedded types follow Go's rules (explicit selector required if ambiguous).

### 7.7 No inheritance

Rugby does **not** support Ruby inheritance semantics.

### 7.8 Special Methods

**String Conversion (`to_s`):**
* `def to_s` compiles to `String() string`
* Automatically satisfies Go's `fmt.Stringer` interface
* `puts user` uses this method automatically

**Equality (`==`):**
* Rugby `==` compiles to `runtime.Equal(a, b)`
* **Custom Equality:** If a class defines `def ==(other)`, it compiles to `Equal(other interface{}) bool`.
* **Runtime Dispatch:** `runtime.Equal` follows this logic:
    1. If `a` has an `Equal(interface{})` method, call `a.Equal(b)`.
    2. Else if `a` and `b` are slices/maps, perform a deep comparison.
    3. Else, use standard Go `==` (identity/primitive equality).
* This allows Rugby to support both semantic equality (user-defined) and safe deep equality for Go types.

### 7.9 The `self` keyword

* `self` refers to the current instance (the method receiver).
* It allows method chaining (`return self`) and disambiguation (`self.name` vs local `name`).
* Compiles to the generated receiver variable name (e.g., `u` in `func (u *User)...`).

```ruby
def with_name(n)
  @name = n
  self
end
```


---

## 8. Interfaces

### 8.1 Declaration

```ruby
interface Speaker
  def speak -> String
end
```

Compiles to:

```go
type Speaker interface { Speak() string }
```

**Important:** Interface methods are **implicitly exported** (uppercase in Go). This is required for cross-package interface satisfaction in Go.

* `def speak` in an interface → `Speak()` in Go
* Interfaces should be marked `pub` to be usable from other packages
* Methods use the exported name transformation regardless of `pub`

### 8.2 Structural conformance

* A class satisfies an interface if it has the required methods (like Go)
* No `implements` keyword required

---

## 9. Visibility & Naming

### 9.1 Core principle

**Inside a Rugby module, everything is usable.**

The `pub` keyword controls what becomes visible when the compiled Go package is imported by other Go (or Rugby) modules.

* `pub` = export to Go (uppercase in output)
* no `pub` = internal to package (lowercase in output)

There is no Ruby-style `private`/`public`.

### 9.2 What can be `pub`

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
  pub def inc!
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

### 9.3 Rugby naming conventions

These are **style rules**, not visibility rules:

* Types (`class`, `interface`) → `CamelCase`
* Functions, methods, variables → `snake_case`
* Capitalization in Rugby **never** controls visibility

### 9.4 Go name generation

Rugby names are rewritten to idiomatic Go names:

| Rugby source | pub? | Go output |
|--------------|------|-----------|
| `def parse_json` | no | `parseJSON` |
| `pub def parse_json` | yes | `ParseJSON` |
| `def user_id` | no | `userID` |
| `pub def user_id` | yes | `UserID` |

### 9.5 Acronym list

To keep output Go-idiomatic, the compiler uses an acronym table.

**Default MVP acronyms:**

`id`, `url`, `uri`, `http`, `https`, `json`, `xml`, `api`, `uuid`, `ip`, `tcp`, `udp`, `sql`, `tls`, `ssh`, `cpu`, `gpu`

Rules:

* Exported: `id` → `ID`, `http` → `HTTP`
* Unexported: first-part acronym `http_*` → `http*`, later-part `*_id` → `*ID`

Configurable via compiler config later; hardcoded for MVP.

### 9.6 Name collision rules

Collisions occur when different Rugby identifiers normalize to the same Go identifier:

* `foo_bar` and `foo__bar`
* `inc` and `inc!`

**Rule:** Collisions are compile-time errors showing:

* Both Rugby names and locations
* The resulting Go name
* Suggested fix

### 9.7 Reserved words

If a Go identifier would be a keyword (`type`, `var`, `func`) or conflict with generated names (`main`, `init`, `NewTypeName`):

* **`pub`**: compile error (force rename)
* **internal**: escape with trailing `_` (e.g., `type` → `type_`)

---

## 10. Go Interop

### 10.1 Imports

```ruby
import net/http
import encoding/json as json
```

* `import a/b` → Go `import "a/b"`
* `import a/b as x` → Go `import x "a/b"`

### 10.2 Calling Go functions

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

### 10.3 Struct fields and methods

* `resp.Body` allowed as-is
* Optional: `resp.body` → `resp.Body` for Go types

### 10.4 Defer

```ruby
defer resp.Body.Close
```

Compiles to:

```go
defer resp.Body.Close()
```

Rule: `defer <callable>` compiles to `defer f()`.

---

## 11. Runtime Package

Rugby provides a Go runtime package (`rugby/runtime`) that gives Ruby-like ergonomics to Go's built-in types. This is what makes Rugby feel like Ruby while compiling to idiomatic Go.

### 11.1 Design principles

* **Wrap, don't reinvent:** Functions wrap Go stdlib, adding Ruby-style APIs
* **Type-safe generics:** Use Go 1.18+ generics for collections
* **Zero-cost when unused:** Only imported when Rugby code uses these methods
* **Predictable mapping:** Each Rugby method maps to a clear runtime function

### 11.2 Package organization

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

### 11.3 Array methods (`Array[T]`)

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

### 11.4 Map methods (`Map[K, V]`)

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

### 11.5 String methods

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

### 11.6 Integer methods

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

### 11.7 Float methods

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

### 11.8 Codegen integration

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

### 11.9 Global functions (kernel)

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

### 11.10 Import generation

The compiler automatically adds `import "rugby/runtime"` when any runtime functions are used. All kernel functions and stdlib methods go through the runtime package for consistency.

---

## 12. Diagnostics

* Type errors and unresolved identifiers are compile-time errors
* Interop mapping errors (e.g., `io.read_all` not found) must show:
  * Rugby source span
  * Intended Go package/type
  * Suggested candidates (`ReadAll`)

---

## 13. MVP Scope

### Parser + AST

* imports
* defs (functions)
* classes with initialize + methods
* if/else/elsif
* while loops
* each blocks

### Type inference

* locals, params, returns
* optional `T?` inference

### Go emitter

* generates `.go` files + prelude
* preserves readable formatting

### Go interop resolver

* package member resolution
* snake_case → CamelCase mapping
