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

### 4.3 Type annotations

Optional—infer when omitted:

```ruby
x : Int = 3
def add(a : Int, b : Int) -> Int
  a + b
end
```

If inference fails, compiler error with location.

### 4.4 Optionals (`T?`)

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

`nil` exists only as the "empty" value for `T?` (and reference types). No implicit nil for non-optional value types.

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

### 5.3 Loops

**While:**

```ruby
while cond
  ...
end
```

Compiles to `for cond { ... }`

**Each (range iteration):**

```ruby
arr.each do |v|
  ...
end
```

Compiles to `for _, v := range arr { ... }`

**Each with index:**

```ruby
arr.each_with_index do |v, i|
  ...
end
```

Compiles to `for i, v := range arr { ... }`

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

### 6.3 Errors

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

### 7.3 Constructors

* `initialize` generates a constructor function
* `User.new(...)` rewrites to constructor call
* If class is `pub`: `NewUser(...) *User`
* If internal: `newUser(...) *User`

### 7.4 Methods and receivers

* `def method` → value receiver (default)
* `def method!` → pointer receiver
* Compiler may upgrade to pointer receiver if mutation detected

### 7.5 Methods with `!`

```ruby
def inc!
  @n += 1
end
```

* `!` is stripped from Go name
* Method uses pointer receiver
* `def inc!` → `func (c *Counter) inc()`
* `pub def inc!` → `func (c *Counter) Inc()`

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

### 7.7 No inheritance

Rugby does **not** support Ruby inheritance semantics.

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

## 11. Collections (Prelude)

Minimal standard prelude that lowers to Go:

* `arr.each { |x| ... }` → range loop
* `arr.map { |x| ... }` → allocate + range loop
* `arr.compact` for `T?` arrays → filter present values
* `s.to_i?` → `strconv.Atoi`

Keep it small; only add features that map cleanly to Go.

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
