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
