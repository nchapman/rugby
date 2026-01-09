Below is a shareable **Rugby MVP Language Spec (v0.1)** aimed at “Ruby-like syntax, Go-like semantics” with **first-class Go interop** and **no metaprogramming**.

---

# Rugby Language Spec (MVP v0.1)

**Goal:** Ruby-ish surface syntax that compiles to idiomatic Go and makes using Go libraries feel native.

## 0. Non-Goals (Hard No’s)

Rugby **must not** support:

* Reopening classes / monkey patching
* `eval`, runtime codegen, reflection-driven magic
* `method_missing`, dynamic dispatch hacks
* Runtime modification of methods, fields, or modules
* Implicit global mutation beyond normal variables

Rugby is **compiled**, **static**, and **predictable**.

---

## 1. Execution & Compilation Model

* Rugby compiles to **Go source** (one or more `.go` files), then uses `go build`.
* Output should be **idiomatic** Go:

  * structs + methods
  * interfaces
  * package-level functions
  * standard error handling patterns
* The compiler may generate helper functions/types (prelude), but should keep them small and transparent.

---

## 2. Files, Modules, Imports

### 2.1 File-to-package mapping

* Each Rugby file belongs to a module (Go package).
* Default package name = module name; configurable via build config.

### 2.2 Imports

Syntax:

```ruby
import net/http
import encoding/json as json
```

Rules:

* `import a/b` maps to Go `import "a/b"`.
* Optional alias: `import a/b as x` maps to Go `import x "a/b"`.

Name resolution:

* Referencing imported package members:

  * Prefer **Go-exported names as-is**: `http.Get`, `json.Marshal`
  * Optional ergonomic aliasing is allowed (see 8.2), but must resolve to the correct Go identifier at compile-time.

---

## 3. Lexical & Basic Syntax

* Comments: `# ...` line comments
* Blocks: `do ... end` or `{ ... }` (compiler can support one first; recommended: `do/end`)
* Strings:

  * `"..."` normal string
  * interpolation allowed: `"hi #{name}"` => `fmt.Sprintf` rewrite (or compile-time concat when trivial)

---

## 4. Types & Type Inference

Rugby is statically typed with inference.

### 4.1 Primitive types

* `Int` -> Go `int`
* `Int64` -> Go `int64`
* `Float` -> Go `float64`
* `Bool` -> Go `bool`
* `String` -> Go `string`
* `Bytes` -> Go `[]byte`

### 4.2 Composite types

* `Array[T]` -> `[]T`
* `Map[K, V]` -> `map[K]V`
* `T?` (optional) -> (see 6)

### 4.3 Annotations

Optional annotation syntax:

```ruby
x : Int = 3
def add(a : Int, b : Int) -> Int
  a + b
end
```

Inference:

* If omitted, infer types from literals and usage.
* If inference fails, compiler error with location.

---

## 5. Variables, Assignment, Control Flow

### 5.1 Variables

* `x = expr` declares (if new) or assigns (if existing).
* Shadowing: allowed in nested scopes; discouraged but permitted with clear rules.

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

* Range iteration:

```ruby
arr.each do |v|
  ...
end
```

Compiles to `for _, v := range arr { ... }`

* Indexed iteration:

```ruby
arr.each_with_index do |v, i|
  ...
end
```

Compiles to `for i, v := range arr { ... }`

* While (optional for MVP):

```ruby
while cond
  ...
end
```

Compiles to `for cond { ... }`

---

## 6. Optionals (`T?`) and “truthy” rules

Rugby supports **Optionals** as a first-class compile-time feature.

### 6.1 Representation

Preferred Go lowering:

* `T?` compiles to `(T, bool)` in returns and temporaries
* Or a generated `Option[T]` struct if generics are allowed in target Go version
  MVP recommendation: `(T, bool)` (Go-idiomatic).

### 6.2 Optional-producing conversions

Example:

```ruby
n = s.to_i?  # Int?
```

Compiles to:

* `n, ok := strconv.Atoi(s)`

### 6.3 Using optionals in conditionals

Rule: `if x` where `x` is `T?` means “present”.
Example:

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

### 6.4 `nil`

* `nil` exists only as the “empty” value for `T?` (and reference types if needed).
* No implicit nil for non-optional value types.

---

## 7. Functions

### 7.1 Defining functions

```ruby
def add(a, b)
  a + b
end
```

Compiles to Go `func add(a T, b T) T { ... }` with inferred types.

### 7.2 Multiple returns (Go-style)

Rugby allows:

```ruby
def parse_int(s) -> (Int, Bool)
  ...
end
```

Compiles to:

```go
func parseInt(s string) (int, bool) { ... }
```

### 7.3 Errors (Go-native)

Rugby adopts Go error patterns explicitly.

Option A (MVP-friendly): surface Go’s `error` directly:

```ruby
def read(path : String) -> (Bytes, error)
  os.read_file(path)
end
```

Optional sugar allowed (later):

* `try` / `?` operators, but only if they lower to explicit error checks. (Not required for MVP.)

---

## 8. Go Interop (Top Priority)

### 8.1 Calling Go functions

* Rugby calls Go packages with dot syntax:

  * `http.Get(url)` -> `http.Get(url)`
* If Rugby uses snake_case, it must resolve:

  * `io.read_all` -> `io.ReadAll` (compile-time mapping)

### 8.2 Identifier mapping (optional feature)

Allow calling Go exported identifiers using Ruby-ish names:

* `read_all` maps to `ReadAll`
* `new_request` maps to `NewRequest`
  Rules:
* Only for imported Go packages/types
* Mapping is **compile-time only**
* Compiler error if ambiguous

### 8.3 Struct fields and methods from Go

* `resp.Body` allowed as-is.
* Optional: `resp.body` maps to `resp.Body` only for Go types (compile-time).
  Recommend for “feels native”: allow both, prefer keeping Go casing in docs.

### 8.4 Defer

Rugby:

```ruby
defer resp.Body.Close
```

Compiles to:

```go
defer resp.Body.Close()
```

Rule: `defer <callable>` must compile to `defer f()`.

---

## 9. Classes (Ruby surface, Go core)

### 9.1 Definition

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

### 9.2 Lowering

* `class User` -> `type User struct { name string; age int }`
* `@name`/`@age` become struct fields

  * default: unexported (`name`, `age`)
  * if Rugby supports `pub`, compile to exported fields (`Name`, `Age`)

### 9.3 Construction

* `initialize` lowers to `NewUser(...) *User` (constructor function)
* `User.new(...)` rewrites to `NewUser(...)`

### 9.4 Methods and receivers

* `def method` -> method with **value receiver** by default
* `def method!` -> method with **pointer receiver**
* Compiler may upgrade value receiver to pointer receiver if mutation detected.

Naming:

* `birthday!` cannot exist in Go identifier; compiler emits `Birthday` (or `BirthdayBang`).
  Recommended rule:
* Strip `!` and use pointer receiver; resolve name conflicts by suffixing `Bang`.

### 9.5 No inheritance (in Ruby sense)

Rugby does **not** support Ruby inheritance semantics.

### 9.6 Embedding (Go composition) via `<`

Rugby:

```ruby
class Service < Logger
  def run
    info("running")
  end
end
```

Meaning: **embed** `Logger` in `Service`.
Go:

```go
type Service struct { Logger }
```

Rules:

* `<` is embedding only.
* No `super` in MVP.

---

## 10. Interfaces (Go-native polymorphism)

### 10.1 Declaring interfaces

```ruby
interface Speaker
  def speak -> String
end
```

Compiles to:

```go
type Speaker interface { Speak() string }
```

### 10.2 Structural conformance

* A class/struct satisfies an interface if it has required methods (like Go).
* No `implements` keyword required.

---

## 11. Collections & Common Methods (Minimal Prelude)

Provide a tiny standard prelude that lowers to Go loops/calls.

MVP collection ops:

* `arr.each { |x| ... }` -> range loop
* `arr.map { |x| ... }` -> allocate + range loop
* `arr.compact` for `T?` arrays -> filter present values
* `s.to_i?` -> `strconv.Atoi`

Keep it small; prefer adding features only when they map cleanly to Go.

---

## 12. Visibility & Packaging

Optional for MVP:

* `pub` for exported names:

```ruby
pub class User
  pub def greet -> String
    ...
  end
end
```

Compiles to exported Go identifiers.

If omitted:

* classes/types can be exported by file/module rule or config.

---

## 13. Determinism & Diagnostics

* Compilation must be deterministic.
* Type errors and unresolved identifiers are compile-time errors.
* Interop mapping errors (e.g., `io.read_all` not found) must show:

  * Rugby source span
  * intended Go package/type
  * suggested candidates (`ReadAll`)

---

## 14. MVP Deliverables (Recommended)

1. Parser + AST for:

   * imports
   * defs
   * classes w/ initialize + methods
   * if/else
   * each blocks
2. Type inference:

   * locals, params, returns
   * optional `T?` inference
3. Go emitter:

   * generates `.go` files + prelude
   * preserves readable formatting
4. Go interop resolver:

   * package member resolution
   * optional snake_case mapping for Go identifiers

## Additions

## Export & Naming

This section defines how Rugby names map to Go names and how `pub` controls what is visible to other Go packages.

---

## 1. Core principle (read this first)

**Inside a Rugby module, everything is usable.**

The `pub` keyword exists **only** to control what becomes visible when the compiled Go package is imported by *other* Go (or Rugby) modules.

> `pub` does **not** affect visibility inside the same Rugby module.

There is no Ruby-style `private` / `public` in MVP.

---

## 2. What `pub` means

* `pub` = **export this symbol to Go**
* no `pub` = **internal to the Go package**

That’s it. One rule.

---

## 3. What can be marked `pub`

### 3.1 Functions

```ruby
def helper(x : Int) -> Int
  x * 2
end

pub def double(x : Int) -> Int
  helper(x)
end
```

* `double` is visible to Go
* `helper` is not
* both are usable inside the same Rugby module

---

### 3.2 Classes

```ruby
class Counter
  def initialize(start : Int = 0)
    @n = start
  end

  def inc!
    @n += 1
  end
end
```

* This class is usable inside Rugby
* It is **not** usable from Go

To export it:

```ruby
pub class Counter
  def initialize(start : Int = 0)
    @n = start
  end

  pub def inc!
    @n += 1
  end
end
```

Rules:

* A class must be `pub` to be usable from Go
* Methods intended for Go must also be `pub`

---

### 3.3 Interfaces

```ruby
pub interface Greeter
  pub def greet(name : String) -> String
end
```

* Interfaces meant for Go **must** be `pub`
* All interface methods must be `pub`

---

## 4. Naming rules in Rugby source

These are **style rules**, not visibility rules.

* Types (`class`, `interface`) → `CamelCase`
* Functions, methods, variables → `snake_case`
* Capitalization in Rugby **never controls visibility**

Examples:

```ruby
class HttpServer
end

def parse_json
end
```

---

## 5. Go name generation

Rugby names are rewritten to idiomatic Go names.

### 5.1 Exported (`pub`) names

* `snake_case` → `CamelCase`
* Acronyms preserved (`id` → `ID`, `http` → `HTTP`)

Examples:

* `pub def parse_json` → `ParseJSON`
* `pub def user_id` → `UserID`

### 5.2 Internal (non-`pub`) names

* `snake_case` → `camelCase`
* Acronyms preserved

Examples:

* `def parse_json` → `parseJSON`
* `def user_id` → `userID`

### 5.3 Acronym list

To keep output Go-idiomatic, the compiler uses an acronym table.

**Default MVP acronyms (lowercase keys):**

* `id`, `url`, `uri`, `http`, `https`, `json`, `xml`, `api`, `uuid`, `ip`, `tcp`, `udp`, `sql`, `tls`, `ssh`, `cpu`, `gpu`

Rules:

* Exported: `id` → `ID`, `http` → `HTTP`
* Unexported: first-part acronym `http_*` → `http*`, later-part acronym `*_id` → `*ID`

The list should be configurable via compiler config later, but hardcoded is fine for MVP.

### 5.4 Name collision rules

Collisions can occur when different Rugby identifiers normalize to the same Go identifier.

Examples:

* `foo_bar` and `foo__bar` (empty segment removal)
* `parse_json` and `parse_JSON` (if you normalize case)
* `inc` and `inc!` (after dropping `!`)

**Rule (MVP):**

Collisions are **compile-time errors** with a clear message showing:

* the two Rugby names and their locations
* the resulting Go name they collide on
* a suggested fix (rename one of them)

(You can later add automatic disambiguation, but errors keep APIs intentional.)

### 5.5 Reserved words

If a normalized Go identifier would be a Go keyword (`type`, `var`, `func`, etc.) or would conflict with required generated names (`main`, `init`, `NewTypeName`), the compiler must:

* either error, or
* apply a deterministic escape (MVP recommendation: **escape** by suffixing `_` for internal names only)

**MVP recommended behavior:**

* For **public API (`pub`)**: error (force explicit rename)
* For **internal**: escape with trailing `_` (e.g. `type` → `type_`)

---

## 6. Methods with `!`

If a method ends in `!`:

```ruby
def inc!
  @n += 1
end
```

Rules:

* `!` is removed in Go name
* Method uses a **pointer receiver**
* Exported if and only if marked `pub`

Examples:

* `def inc!` → `func (c *Counter) inc()`
* `pub def inc!` → `func (c *Counter) Inc()`

---

## 7. Constructors

* `initialize` generates a constructor
* If the class is `pub`, the constructor is `NewTypeName`
* Otherwise it is internal

Example:

```ruby
pub class Counter
  def initialize(n : Int)
    @n = n
  end
end
```

→

```go
func NewCounter(n int) *Counter
```

---

## 8. Simple restriction (compiler-enforced)

To avoid useless exports:

* ❌ `pub def` inside a non-`pub` class is an error

Reason: Go callers cannot name the type anyway.

---

## 9. Summary (one screen)

* Everything is usable inside a Rugby module
* `pub` only controls Go export
* Capitalization in Rugby has no visibility meaning
* Types are `CamelCase`, functions are `snake_case`
* Exported Go names are `CamelCase`
* Internal Go names are `camelCase`
