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
| `Bytes`  | `[]byte`  |

### 4.2 Composite types

| Rugby        | Go          |
|--------------|-------------|
| `Array[T]`   | `[]T`       |
| `Map[K, V]`  | `map[K]V`   |
| `T?`         | `(T, bool)` |
| `Range`      | `struct` (internal) |

### 4.2.1 Range Literals

* `start..end` → inclusive (0..5 includes 5)
* `start...end` → exclusive (0...5 excludes 5)
* Primarily used in `for` loops.

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

**Status:** Planned for future phase (after basic type annotations).

Representation (MVP): `T?` compiles to `(T, bool)`.

```ruby
n = s.to_i?  # Int?
```

Compiles to: `n, ok := strconv.Atoi(s)`

Using in conditionals—`if x` where `x` is `T?` means "present":

```ruby
if (n = s.to_i?)
  puts n + 1
end
```

Compiles to:

```go
if n, ok := strconv.Atoi(s); ok {
    fmt.Println(n + 1)
}
```

**Reference types:** For reference types (pointers, slices, maps, interfaces), `T?` is **disallowed** in MVP. These types already have `nil` semantics in Go.

Examples:
* `Int?` → valid (`(int, bool)`)
* `String?` → valid (`(string, bool)`)
* `*User?` → **compile error** (use `*User` which can already be nil)
* `Array[Int]?` → **compile error** (use `[]int` which can already be nil)

`nil` is reserved for direct use with reference types only.

---

## 5. Variables & Control Flow

### 5.1 Variables

* `x = expr` declares (if new) or assigns (if existing)
* Shadowing allowed in nested scopes

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

Compiles to Go `if/else if/else`.

### 5.3 Imperative Loops (Statements)

Use loops when you need **control flow** (searching, early exit, side effects). These are statements, not expressions (they do not return a value).

**For Loop:**
The primary way to iterate with control flow.

```ruby
for item in items
  return item if item.id == 5
end

for i in 0..10
  puts i
end
```

Compiles to:
* Collections: `for _, item := range items { ... }`
* Ranges: `for i := 0; i <= 10; i++ { ... }`

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
* **Return:** `return` exits the **block only** (returns a value to the iterator), *not* the enclosing function.
* **Break:** `break` is **not supported** inside blocks (use `for` if you need to stop early).

```ruby
# Data transformation (Expression)
names = users.map do |u|
  return "Guest" if u.nil?  # Returns "Guest" to the map array, iteration continues
  u.name                    # Implicit return
end
```

**Compilation:**
All methods with blocks compile to runtime calls receiving a function literal:

```go
// Ruby: users.each { |u| puts u }
runtime.Each(users, func(u User) {
    fmt.Println(u)
})
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

Instance variable types are **inferred from `initialize` only** (MVP):

```ruby
class User
  def initialize(name : String, age : Int)
    @name = name  # @name inferred as String
    @age = age    # @age inferred as Int
  end
end
```

Rules:
* All instance variables must be assigned in `initialize`
* Using `@var` before assignment in `initialize` is a compile error
* Types are inferred from the assigned expression
* Optional: explicit type annotations on instance variables (post-MVP)

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
class Service < Logger
  def run
    info("running")
  end
end
```

Compiles to:

```go
type Service struct { Logger }
```

* `<` means embedding only (not inheritance)
* No `super` in MVP
* Multiple embedding not supported in MVP (single embed only)

### 7.7 No inheritance

Rugby does **not** support Ruby inheritance semantics.

### 7.8 Special Methods

**String Conversion (`to_s`):**
* `def to_s` compiles to `String() string`
* Automatically satisfies Go's `fmt.Stringer` interface
* `puts user` uses this method automatically

**Equality (`==`):**
* Rugby `==` compiles to `runtime.Equal(a, b)` for non-primitive types
* Handles slice/map equality (deep comparison)
* To customize equality for a class, define `def ==(other)` (future phase)

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
