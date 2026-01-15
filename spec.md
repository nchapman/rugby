# Rugby Language Specification

**Goal:** Provide a robust, Ruby-inspired surface syntax that compiles to idiomatic Go, enabling developers to build high-performance systems with Go's power and Ruby's elegance.

Rugby is **compiled**, **statically typed**, and **deterministic**.

---

## 1. Design Philosophy

Rugby occupies a unique position: Ruby's expressiveness meets Go's performance and simplicity. The language makes deliberate trade-offs:

| Ruby | Rugby | Go |
|------|-------|-----|
| Dynamic typing | Static typing with inference | Static typing |
| Duck typing | Structural interfaces | Structural interfaces |
| Metaprogramming | No metaprogramming | No metaprogramming |
| GC + interpreter | GC + native compilation | GC + native compilation |
| Blocks everywhere | Lambdas for iteration | First-class functions |
| Exceptions | Errors as values | Errors as values |

**Core principles:**

1. **Readable output** - Generated Go should be idiomatic and debuggable
2. **No magic** - Every construct has a clear, predictable compilation
3. **Ruby comfort** - Familiar syntax for Ruby developers
4. **Go power** - Full access to Go's ecosystem and performance
5. **Fail early** - Catch errors at compile time, not runtime

---

## 2. Non-Goals

Rugby **must not** support:

* Reopening classes / monkey patching
* `eval`, runtime codegen, reflection-driven magic
* `method_missing`, dynamic dispatch hacks
* Runtime modification of methods, fields, or modules
* Implicit global mutation beyond normal variables

These constraints enable static analysis, predictable performance, and readable generated code.

---

## 3. Compilation Model

* Rugby compiles to **Go source** (one or more `.go` files), then uses `go build`
* Output should be **idiomatic** Go:
  * structs + methods
  * interfaces
  * package-level functions
  * standard error handling patterns
* The compiler may generate helper functions/types (prelude), but should keep them small and transparent
* Compilation must be deterministic

### 3.1 Bare Scripts (Top-Level Execution)

Rugby files can contain executable statements at the top level without an explicit `main` function.

```ruby
# hello.rg
puts "Hello, world!"

# With functions - definitions are lifted, calls execute in order
def greet(name : String)
  puts "Hello, #{name}!"
end

greet("Rugby")
```

**Rules:**
- Top-level statements execute in source order
- Only one file per directory may contain top-level statements
- `def main` and top-level statements are mutually exclusive (compile error if both)
- Definitions (`def`, `class`, `interface`) become package-level constructs

| File contains | Mode | Output |
|---------------|------|--------|
| Only defs/classes | Library | `package <name>`, no main |
| Top-level statements | Script | `package main` + generated `func main()` |
| `def main` | Explicit | `package main` + user's `func main()` |

### 3.2 Multi-File Projects

Rugby projects follow Go's package model:

```
myapp/
  main.rg           # package main, has top-level statements
  util.rg           # package main, only definitions
  lib/
    parser.rg       # package parser
    lexer.rg        # package lexer
```

**Rules:**
- All `.rg` files in a directory belong to the same package
- Package name derived from directory name (or `main` for executables)
- Circular imports are a compile error
- Import paths mirror Go's module system

### 3.3 Initialization Order

Within a file:
1. Constants are evaluated first (compile-time where possible)
2. Package-level variables are initialized in declaration order
3. Top-level statements execute in source order

Across files in a package:
- Files are processed in lexicographic order
- Dependencies are resolved automatically (forward references allowed)

---

## 4. Lexical Structure

### 4.1 Comments

```ruby
# single line comment
```

Block comments are not supported (use multiple `#` lines).

### 4.2 Identifiers

```
identifier     = letter { letter | digit | "_" }
type_name      = uppercase_letter { letter | digit }
predicate      = identifier "?"
setter         = identifier "="
instance_var   = "@" identifier
```

**Naming conventions:**
- Types: `CamelCase` (`User`, `HttpClient`)
- Functions/methods: `snake_case` (`find_user`, `to_json`)
- Predicates: `snake_case?` (`empty?`, `valid?`) - must return `Bool`
- Setters: `snake_case=` (`name=`, `value=`)
- Constants: `SCREAMING_SNAKE_CASE` (`MAX_SIZE`, `DEFAULT_PORT`)
- Instance variables: `@snake_case` (`@name`, `@user_id`)

### 4.3 Lambdas

```ruby
-> (x) { x * 2 }       # single-line lambda
-> (x) do              # multiline lambda
  x * 2
end
```

See Section 10.6 for full lambda documentation.

### 4.4 Strings

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

### 4.5 Heredocs

Multi-line string literals using `<<` syntax:

```ruby
text = <<END
First line
Second line
END

# <<~ strips common leading whitespace (most useful in practice)
def query
  sql = <<~SQL
    SELECT *
    FROM users
    WHERE active = true
  SQL
end
# sql = "SELECT *\nFROM users\nWHERE active = true"
```

| Syntax | Behavior |
|--------|----------|
| `<<DELIM` | Closing delimiter must be at line start |
| `<<-DELIM` | Closing delimiter can be indented |
| `<<~DELIM` | Strips common leading whitespace from content |

String interpolation (`#{}`) works within heredocs. The delimiter can be any identifier (`END`, `SQL`, `HTML`, etc.).

### 4.6 Symbols

Symbols (`:ok`, `:status`, `:not_found`) are compile-time string constants:

```ruby
status = :ok
state = :pending

:foo == :foo  # true
```

Symbols compile to Go `string` constants. Use symbols for fixed identifiers like status codes, keys, and tags.

### 4.7 Regular Expressions

```ruby
pattern = /\d+/
email_re = /\A[\w+\-.]+@[a-z\d\-]+\.[a-z]+\z/i

if text =~ /hello/
  puts "matched"
end

# Named captures
if match = text.match(/(?P<name>\w+)@(?P<domain>\w+)/)
  puts match["name"]
end
```

| Operator | Description |
|----------|-------------|
| `=~` | Returns true if string matches pattern |
| `!~` | Returns true if string does not match |
| `.match(re)` | Returns `Match?` with captures |
| `.scan(re)` | Returns `Array<Match>` of all matches |
| `.gsub(re, replacement)` | Replace all matches |
| `.sub(re, replacement)` | Replace first match |

Regular expressions compile to Go's `regexp.Regexp`. The `/.../` literal compiles the pattern at package init time.

Flags: `i` (case-insensitive), `m` (multiline), `x` (extended/verbose)

---

## 5. Types

Rugby is statically typed with inference.

### 5.1 Primitive Types

| Rugby    | Go        | Notes |
|----------|-----------|-------|
| `Int`    | `int`     | Platform-dependent size |
| `Int8`   | `int8`    | |
| `Int16`  | `int16`   | |
| `Int32`  | `int32`   | |
| `Int64`  | `int64`   | |
| `UInt`   | `uint`    | |
| `UInt8`  | `uint8`   | Also `Byte` |
| `UInt16` | `uint16`  | |
| `UInt32` | `uint32`  | |
| `UInt64` | `uint64`  | |
| `Float`  | `float64` | |
| `Float32`| `float32` | |
| `Bool`   | `bool`    | |
| `String` | `string`  | Immutable UTF-8 |
| `Symbol` | `string`  | Interned constant |
| `Bytes`  | `[]byte`  | Mutable byte slice |
| `Rune`   | `rune`    | Unicode code point |

### 5.2 Composite Types

| Rugby        | Go          | Notes |
|--------------|-------------|-------|
| `Array<T>`   | `[]T`       | Dynamic array (slice) |
| `Map<K, V>`  | `map[K]V`   | Hash map |
| `Set<T>`     | `map[T]struct{}` | Unique values |
| `(T, U)`     | `struct` or multiple returns | Tuple type |
| `T?`         | `(T, bool)` or `runtime.Option[T]` | Optional (tuple for returns, struct for storage) |
| `Range`      | `struct`    | Numeric range |
| `Chan<T>`    | `chan T`    | Channel |
| `Task<T>`    | `struct`    | Async result |
| `Regex`      | `*regexp.Regexp` | Compiled pattern |
| `Void`       | (no return) | Unit type for side-effect functions |

### 5.3 Type Aliases

```ruby
type UserID = Int64
type Handler = (Request) -> Response
type StringMap = Map<String, String>
```

Type aliases are transparent - `UserID` and `Int64` are interchangeable.

### 5.4 Tuples

Tuples are fixed-size, heterogeneous collections used primarily for multiple return values:

```ruby
# Tuple types in function signatures
def parse(s : String) -> (Int, Bool)
  # ...
end

def get_bounds -> (Int, Int, Int, Int)
  0, 0, width, height
end

# Destructuring (required for access)
n, ok = parse("42")
x, y, w, h = get_bounds()
first, _ = get_pair()  # ignore second element

# Typed tuple variables
coords : (Int, Int) = get_point()
```

**Rules:**
- Tuple size is part of the type: `(Int, Int)` and `(Int, Int, Int)` are different types
- Elements are accessed via destructuring only (no index access)
- Tuples are immutable value types
- Compiles to Go multiple return values or generated structs

### 5.5 Array Literals

```ruby
nums = [1, 2, 3]
words = %w{foo bar baz}        # ["foo", "bar", "baz"]
words = %W{hello #{name}}      # interpolated: ["hello", "world"]
all = [1, *rest, 4]            # splat: [1, 2, 3, 4]
empty : Array<Int> = []        # typed empty array
```

Word arrays support any delimiter: `%w{...}`, `%w(...)`, `%w[...]`, `%w<...>`, `%w|...|`

**Type inference for arrays:**
```ruby
[1, 2, 3]           # Array<Int>
["a", "b"]          # Array<String>
[1, "two", 3.0]     # Error: mixed types require explicit Array<any>
[1, "two"] : Array<any>  # OK: explicit any
```

### 5.6 Indexing

Negative indices count from the end:

```ruby
arr[-1]   # last element
arr[-2]   # second-to-last
str[-1]   # last character (Unicode-aware)
```

Out-of-bounds indices panic. For safe access, use `first`/`last` which return optionals.

### 5.7 Range Slicing

```ruby
arr = ["a", "b", "c", "d", "e"]
arr[1..3]   # ["b", "c", "d"] (indices 1, 2, 3 inclusive)
arr[1...3]  # ["b", "c"] (indices 1, 2, exclusive end)
arr[-3..-1] # ["c", "d", "e"] (last 3 elements)
arr[2..]    # ["c", "d", "e"] (from index 2 to end)
arr[..2]    # ["a", "b", "c"] (from start through index 2)

str = "hello"
str[1..3]   # "ell" (characters at indices 1, 2, 3)
```

Out-of-bounds ranges return empty. Negative indices count from end.

### 5.8 Map Literals

```ruby
m = {"name" => "Alice", "age" => 30}  # hash rocket
m = {name: "Alice", age: 30}          # symbol shorthand (same result)
m = {name:, age:}                     # implicit value (uses variables)
m = {**defaults, **overrides}         # double splat merges maps
empty : Map<String, Int> = {}         # typed empty map
```

**Braces disambiguation:**

Since lambdas use arrow syntax (`-> { }`), bare braces are always map literals:

```ruby
foo({a: 1})          # Map argument (explicit parens)
foo bar, {a: 1}      # Map argument (comma separates args)
foo(bar, {a: 1})     # Explicit: map as second arg to foo

# Lambdas use arrow syntax
foo -> { expr }      # Lambda argument
foo -> (x) { x * 2 } # Lambda with parameter
```

When in doubt, use parentheses to be explicit.

### 5.9 Set Literals

```ruby
s = Set{1, 2, 3}
s = Set<String>{"a", "b"}

s.include?(1)  # true
s.add(4)
s.delete(1)
s | other      # union
s & other      # intersection
s - other      # difference
```

### 5.10 Range Type

```ruby
1..5    # inclusive: 1, 2, 3, 4, 5
1...5   # exclusive: 1, 2, 3, 4

for i in 0..5
  puts i
end

(1..100).include?(50)
nums = (1..5).to_a  # [1, 2, 3, 4, 5]
```

Methods: `each`, `include?`/`contains?`, `to_a`, `size`/`length`

Ranges are integer-only (`Int`, `Int64`, etc.). For other sequences, use explicit iteration or arrays.

### 5.11 Channel Type

```ruby
ch = Chan<Int>.new(10)  # buffered
ch = Chan<Int>.new      # unbuffered
```

See Section 17 for channel operations.

### 5.12 Task Type

```ruby
t = spawn { expensive_work() }  # Task<Result>
result = await t
```

See Section 17 for spawning and awaiting.

---

## 6. Generics

Rugby supports generic types and functions.

### 6.1 Generic Functions

```ruby
def identity<T>(x : T) -> T
  x
end

def swap<T, U>(a : T, b : U) -> (U, T)
  b, a
end

def map_values<K, V, R>(m : Map<K, V>, f : (V) -> R) -> Map<K, R>
  result = Map<K, R>{}
  for k, v in m
    result[k] = f.(v)
  end
  result
end
```

### 6.2 Generic Classes

```ruby
class Box<T>
  def initialize(@value : T)
  end

  def map<R>(f : (T) -> R) -> Box<R>
    Box<R>.new(f(@value))
  end
end

box = Box<Int>.new(42)
str_box = box.map -> (n) { n.to_s }
```

### 6.3 Generic Interfaces

```ruby
interface Container<T>
  def get -> T?
  def put(value : T)
end
```

### 6.4 Type Constraints

```ruby
# Constrain to types implementing an interface
def sort<T : Comparable>(arr : Array<T>) -> Array<T>
  # ...
end

# Multiple constraints
def process<T : Readable & Closeable>(resource : T)
  data = resource.read
  resource.close
end

# Built-in constraints
def sum<T : Numeric>(values : Array<T>) -> T
  values.reduce(T.zero) -> (acc, v) { acc + v }
end
```

Built-in constraints:
- `Numeric` - `Int`, `Int64`, `Float`, etc.
- `Ordered` - Types supporting `<`, `>`, `<=`, `>=`
- `Equatable` - Types supporting `==`, `!=`
- `Hashable` - Types usable as map keys

### 6.5 Type Inference for Generics

```ruby
# Type parameters inferred from arguments
identity(42)        # T inferred as Int
identity("hello")   # T inferred as String

# Explicit when needed
Box<Int>.new(42)
result : Array<String> = []
```

---

## 7. Enums

Rugby supports simple enums that compile to Go integer constants.

### 7.1 Basic Enums

```ruby
enum Status
  Pending
  Active
  Completed
  Cancelled
end

status = Status::Active

case status
when Status::Pending
  puts "waiting"
when Status::Active
  puts "running"
else
  puts "done"
end
```

Enums compile to Go `int` constants with a generated `String()` method.

### 7.2 Enums with Explicit Values

```ruby
enum HttpStatus
  Ok = 200
  Created = 201
  BadRequest = 400
  NotFound = 404
  ServerError = 500
end

status = HttpStatus::NotFound
puts status.value  # 404
```

### 7.3 Enum Methods

```ruby
enum Color
  Red
  Green
  Blue
end

Color::Red.to_s     # "Red"
Color.values        # [Color::Red, Color::Green, Color::Blue]
Color.from_string("Red")  # Color::Red (returns Color?)
```

**Note:** Rugby does not support data-carrying enum variants (sum types). Use `T?` for optional values and `(T, error)` for fallible operations—these patterns cover most use cases and map directly to idiomatic Go.

---

## 8. Optionals (`T?`)

`T?` represents optional values. Compiles to `(T, bool)` for returns, `runtime.Option<T>` for storage.

### 8.1 Optional Operators

| Operator | Description |
|----------|-------------|
| `??` | Nil coalescing: `expr ?? default` |
| `&.` | Safe navigation: `obj&.method` (returns `R?`) |

```ruby
name = find_user(id)&.name ?? "Anonymous"
city = user&.address&.city ?? "Unknown"
port = config.get("port") ?? 8080
```

**Rules:**
* `??` requires left side to be `T?` and right side to be `T`
* `&.` requires left side to be `T?`, returns `R?`
* Using `??` or `&.` on non-optional types is a compile error

**Important: `??` checks presence, not truthiness:**
```ruby
# ?? only checks if value is present, not if it's "truthy"
count : Int? = 0
count ?? 10        # Returns 0 (0 IS present)

flag : Bool? = false
flag ?? true       # Returns false (false IS present)

name : String? = ""
name ?? "default"  # Returns "" (empty string IS present)

missing : Int? = nil
missing ?? 10      # Returns 10 (no value present)
```

Assigning a value to an optional makes it present; assigning `nil` makes it absent. This differs from Ruby's `||` which treats `nil`, `false`, and sometimes `0` as falsy.

### 8.2 Optional Methods

| Method | Description |
|--------|-------------|
| `ok?` / `present?` | Is value present? |
| `nil?` / `absent?` | Is value absent? |
| `unwrap` | Unwrap, panic if absent |
| `unwrap_or(default)` | Unwrap or return default |
| `map -> (x) { }` | Transform if present -> `R?` |
| `flat_map -> (x) { }` | Transform if present, flatten -> `R?` |
| `each -> (x) { }` | Execute lambda if present |
| `filter -> (x) { }` | Keep if predicate true -> `T?` |

### 8.3 Unwrapping Patterns

```ruby
# if let - scoped binding (preferred)
if let user = find_user(id)
  puts user.name
end

# if let with else
if let user = find_user(id)
  puts user.name
else
  puts "not found"
end

# Tuple unpacking - Go's comma-ok idiom
user, ok = find_user(id)
if ok
  puts user.name
end

# Guard clause with ??
user = find_user(id) ?? return nil
```

**Scoping rules for `if let`:**
```ruby
user = User.new("Outer", 30)

if let user = find_user(id)  # New binding, shadows outer `user`
  puts user.name             # Refers to found user
end

puts user.name  # Refers to outer user ("Outer") - shadowing ends at block
```

Variables bound in `if let` are scoped to the `if` block and do not affect outer variables.

| Pattern | Use case |
|---------|----------|
| `expr ?? default` | Fallback value |
| `expr&.method` | Safe navigation |
| `if let u = expr` | Conditional with binding |
| `val, ok = expr` | Need boolean separately |
| `expr.unwrap` | Absence is a bug (panics) |

### 8.4 The `nil` Keyword

`nil` is valid only for `T?` (optional) and `error` types:

```ruby
def find_user(id : Int) -> User?
  return nil if id < 0
  User.new("Alice")
end

x : Int = nil  # ERROR: use Int?

err = save()
if err == nil
  puts "success"
end
```

---

## 9. Variables, Constants, and Control Flow

### 9.1 Variables

```ruby
x = expr       # declares (new) or assigns (existing)
x : Int = 5    # explicit type annotation
x += 5         # compound: +=, -=, *=, /=, %=, &=, |=, ^=
x ||= default  # assign if absent (x must be T?)
```

**`||=` semantics:**
```ruby
name : String? = nil
name ||= "Anonymous"  # assigns "Anonymous" (was absent)

name : String? = ""
name ||= "Anonymous"  # keeps "" (empty string IS present)
```

`||=` only works on optional types (`T?`). For boolean logic, use explicit assignment: `x = x || default`.

Variables are mutable by default. Rugby does not have a `let` vs `var` distinction.

### 9.2 Constants

```ruby
const MAX_SIZE = 1024
const DEFAULT_HOST = "localhost"
const VALID_STATES = %w{pending active done}
const PI = 3.14159

# Typed constants
const TIMEOUT : Int64 = 30

# Computed constants (must be compile-time evaluable)
const BUFFER_SIZE = MAX_SIZE * 2
```

Constants:
- Must be initialized at declaration
- Cannot be reassigned
- Must be compile-time evaluable (literals, other constants, basic arithmetic)
- Conventionally `SCREAMING_SNAKE_CASE`

### 9.3 Destructuring

```ruby
# Tuple destructuring
a, b = get_pair()
first, *rest = items
*head, last = items
_, second, _ = triple

# Map destructuring
{name:, age:} = user_data
{name: n, age: a} = user_data  # rename

# In function parameters
def process({name:, age:} : UserData)
  puts "#{name} is #{age}"
end
```

### 9.4 Conditionals

Conditions must be `Bool` - no implicit truthiness:

```ruby
if valid?           # OK
if user             # ERROR: User is not Bool
if user.ok?         # OK

unless valid?       # inverse of if
  puts "invalid"
end

if let user = find_user(id)   # unwrap optional
  puts user.name
end

if let w = obj.as(Writer)     # type assertion
  w.write("hello")
end

# case - value matching
case status
when 200, 201
  puts "ok"
when 400..499
  puts "client error"
when 500..599
  puts "server error"
else
  puts "unknown"
end

# ternary (right-associative)
max = a > b ? a : b
```

### 9.5 `case` Expressions

Use `case` for value matching:

```ruby
case status
when 200, 201
  puts "ok"
when 400..499
  puts "client error"
when 500..599
  puts "server error"
else
  puts "unknown"
end

# With enums
case color
when Color::Red
  "stop"
when Color::Green
  "go"
when Color::Yellow
  "caution"
end
```

### 9.6 `case_type` for Type Matching

Use `case_type` for type-based dispatch (compiles to Go type switch):

```ruby
case_type obj
when s : String
  "string: #{s}"
when n : Int
  "int: #{n}"
when arr : Array<Int>
  "int array of #{arr.length}"
else
  "unknown type"
end
```

`case_type` requires the value to be an interface type (`any` or a defined interface).

### 9.7 Loops

```ruby
for item in items        # iterate collection
  next if item.skip?
  break if item.done?
  process(item)
end

for i in 0..10           # iterate range
  puts i
end

for key, value in map    # iterate map
  puts "#{key}: #{value}"
end

while running
  process_next()
end

until queue.empty?       # while NOT condition
  handle(queue.pop)
end

loop do                  # infinite loop
  break if done?
  process()
end
```

Control: `break` (exit loop), `next` (skip to next iteration), `return` (exit function)

### 9.8 Statement Modifiers

```ruby
return nil if id < 0
puts "error" unless valid?
break if done
next unless item.valid?
```

### 9.9 Loop Modifiers

```ruby
puts items.shift while items.any?
process(queue.pop) until queue.empty?
```

Unlike statement modifiers (`if`/`unless`), loop modifiers execute repeatedly.

---

## 10. Functions

### 10.1 Basic Functions

```ruby
def add(a, b)
  a + b                          # implicit return (last expression)
end

def add(a : Int, b : Int) -> Int # explicit types
  a + b
end

def find_user(id) -> User?
  return nil if id < 0           # explicit early return
  users[id]
end
```

### 10.2 Multiple Return Values

```ruby
def parse(s) -> (Int, Bool)
  return 0, false if s.empty?
  s.to_i, true
end

# Destructuring
n, ok = parse("42")     # both values
n, _ = parse("42")      # ignore second with _
```

### 10.3 Default Parameters

```ruby
def connect(host : String, port : Int = 8080, timeout : Int = 30)
  # ...
end

connect("localhost")              # port=8080, timeout=30
connect("localhost", 3000)        # timeout=30
connect("localhost", 3000, 60)
```

Default values must be compile-time constants or literals.

### 10.4 Named Parameters

```ruby
def fetch(url : String, headers: Map<String, String> = {},
          timeout: Int = 30, retry: Bool = true)
  # ...
end

fetch("https://example.com")
fetch("https://example.com", timeout: 60)
fetch("https://example.com", retry: false, headers: {"Auth" => token})
```

**Rules:**
- Named parameters come after positional parameters
- Named parameters can be passed in any order
- Mixing positional and named: positional first, then named

### 10.5 Variadic Functions

```ruby
def log(level : Symbol, *messages : String)
  for msg in messages
    puts "[#{level}] #{msg}"
  end
end

log(:info, "Starting", "Loading config", "Ready")

def format(template : String, *args : any) -> String
  sprintf(template, *args)
end

# Splat to pass array as variadic
items = ["a", "b", "c"]
log(:debug, *items)
```

### 10.6 Lambdas

Lambdas are anonymous functions. Use `{ }` for single-line and `do...end` for multiline:

```ruby
# Single-line
double = -> (x : Int) { x * 2 }
users.map -> (u) { u.name }

# Multiline
process = -> (data : String) do
  parsed = parse(data)
  validate(parsed)
  transform(parsed)
end

users.each -> (u) do
  puts u.name
end

# With explicit return type
handler = -> (req : Request) -> Response do
  process(req)
end

# Multiple parameters
map.each -> (k, v) { puts "#{k}: #{v}" }
```

**Calling lambdas:**
```ruby
double = -> (x) { x * 2 }
result = double.(5)     # 10 (preferred)
result = double.call(5) # 10 (alternative)
```

Both `.()` and `.call()` compile to the same Go code. Use `.()` for conciseness.

**Lambda rules:**
- `return` returns from the lambda (not the enclosing function)
- Use `for` loops when you need `break`, `next`, or early function return

```ruby
# Lambda iteration - return skips to next item
users.each -> (u) do
  return if u.inactive?  # skips this user, continues iteration
  process(u)
end

# For loop - full control flow
for user in users
  return user if user.admin?  # returns from enclosing function
  next if user.inactive?      # skips to next iteration
  break if user.done?         # exits loop
end
```

**Passing lambdas to functions:**
```ruby
def with_retry(attempts : Int, action : () -> Bool) -> Bool
  for i in 0...attempts
    return true if action.()
  end
  false
end

with_retry(3, -> { fetch_data() })
```

### 10.7 Closures

Lambdas capture variables from their enclosing scope:

```ruby
def make_counter -> () -> Int
  count = 0
  -> { count += 1 }
end

counter = make_counter()
puts counter.()  # 1
puts counter.()  # 2
puts counter.()  # 3
```

**Capture semantics:**
- Variables are captured by reference
- Mutations inside closures affect the original variable
- Captured variables extend their lifetime

```ruby
# Capture by reference (default)
multiplier = 2
scale = -> (x : Int) { x * multiplier }
scale.(5)       # 10
multiplier = 3
scale.(5)       # 15

# Common gotcha in loops - use explicit capture
funcs = Array<() -> Int>{}
for i in 0..3
  i_copy = i  # capture current value
  funcs << -> { i_copy }
end
```

### 10.8 Function Types

```ruby
type Predicate<T> = (T) -> Bool
type Transform<T, R> = (T) -> R
type Callback = () -> Void
type Handler = (Request) -> (Response, error)

def filter<T>(items : Array<T>, pred : Predicate<T>) -> Array<T>
  result = Array<T>{}
  for item in items
    result << item if pred.(item)
  end
  result
end
```

### 10.9 Calling Convention

Parentheses are optional (command syntax):

```ruby
puts "hello"       # puts("hello")
add 1, 2           # add(1, 2)
foo.bar baz        # foo.bar(baz)
```

**Key rules:**
- Arguments end at newline, `do`, or `and`/`or`
- `foo -1` -> `foo(-1)`, `foo - 1` -> binary subtraction
- `foo [1]` -> array arg, `foo[1]` -> indexing
- Maps need parens: `foo({a: 1})` (braces are for map literals and lambdas)

---

## 11. Classes

### 11.1 Class Definition

```ruby
class User
  @email : String               # explicit field declaration

  def initialize(@name : String, @age : Int)
    @email = "unknown"          # inferred field
  end

  def greet -> String
    "Hello, #{@name}"
  end
end

u = User.new("Alice", 30)
```

### 11.2 Instance Variables

Instance variables are introduced via:
1. **Explicit declaration:** `@field : Type` at class level
2. **Parameter promotion:** `def initialize(@field : Type)`
3. **First assignment in `initialize`**

```ruby
class User
  @role : String                # explicit declaration

  def initialize(@name : String) # parameter promotion
    @age = 0                     # inferred from initialize
  end
end
```

**Rules:**
- New instance variables may only be introduced in `initialize`
- Non-optional fields must be assigned on all control-flow paths
- Within methods, `self` refers to the receiver
- Methods use pointer receivers in generated Go

### 11.3 Accessors

```ruby
getter age : Int        # declares @age + generates def age -> Int
setter name : String    # declares @name + generates def name=(v)
property email : String # both getter and setter

# With visibility
pub getter id : Int
pub property name : String
```

**Field access:**
- Inside class methods: use `@field`
- Outside class: access via methods only (no direct field access)
- Use `pub getter`/`pub property` to export accessors

### 11.4 Custom Accessors

```ruby
class Temperature
  def initialize(@celsius : Float)
  end

  getter celsius : Float

  def fahrenheit -> Float
    @celsius * 9 / 5 + 32
  end

  def fahrenheit=(f : Float)
    @celsius = (f - 32) * 5 / 9
  end
end
```

### 11.5 Class Methods

```ruby
class User
  @@count = 0  # class variable

  def initialize(@name : String)
    @@count += 1
  end

  def self.count -> Int
    @@count
  end

  def self.create(name : String) -> User
    User.new(name)
  end
end

puts User.count
user = User.create("Alice")
```

**Class variable rules:**
- Class variables (`@@var`) are shared across all instances of the class
- Subclasses inherit and share the same class variables
- Class variables are package-private (not exported)

### 11.6 Inheritance

```ruby
class Animal
  def initialize(@name : String)
  end

  def speak -> String
    "..."
  end
end

class Dog < Animal
  def initialize(name : String, @breed : String)
    super(name)
  end

  def speak -> String
    "Woof!"
  end
end

dog = Dog.new("Rex", "Labrador")
dog.speak  # "Woof!"
```

**Rules:**
- Single inheritance only (`class Child < Parent`)
- `super` calls parent's implementation
- `super` with no args forwards all arguments
- `super(args)` passes specific arguments
- Concrete types are invariant (use interfaces for polymorphism)

### 11.7 Designing for Abstraction

Rugby does not have abstract classes. Use interfaces for abstraction and modules for shared implementation:

```ruby
# Define the contract
interface Shape
  def area -> Float
  def perimeter -> Float
end

# Shared behavior via module
module Describable
  def describe -> String
    "Area: #{area}, Perimeter: #{perimeter}"
  end
end

# Concrete implementations
class Circle implements Shape
  include Describable

  def initialize(@radius : Float)
  end

  def area -> Float
    Math::PI * @radius ** 2
  end

  def perimeter -> Float
    2 * Math::PI * @radius
  end
end
```

This pattern separates interface (what) from implementation (how), which aligns with Go's design philosophy.

---

## 12. Structs

Structs are lightweight value types for data without behavior.

### 12.1 Struct Definition

```ruby
struct Point
  x : Int
  y : Int
end

p = Point{x: 10, y: 20}
puts p.x  # 10
```

### 12.2 Struct Features

Structs automatically provide:
- All-fields constructor
- Field accessors (getters)
- Value equality (`==`)
- Hash code for use as map keys
- String representation

```ruby
struct User
  id : Int64
  name : String
  email : String
end

u1 = User{id: 1, name: "Alice", email: "alice@example.com"}
u2 = User{id: 1, name: "Alice", email: "alice@example.com"}

u1 == u2  # true (value equality)

users = Map<User, Int>{}  # structs can be map keys
```

### 12.3 Struct Methods

Structs can have methods, but they receive value (not pointer) receivers:

```ruby
struct Point
  x : Int
  y : Int

  def distance_to(other : Point) -> Float
    dx = (other.x - @x).to_f
    dy = (other.y - @y).to_f
    Math.sqrt(dx ** 2 + dy ** 2)
  end

  def translate(dx : Int, dy : Int) -> Point
    Point{x: @x + dx, y: @y + dy}  # returns new struct
  end
end
```

### 12.4 Struct Immutability

Structs are **immutable by design**. Attempting to modify a field is a compile error:

```ruby
struct Point
  x : Int
  y : Int

  def try_move(dx : Int)
    @x += dx  # ERROR: cannot modify struct field
  end

  def moved(dx : Int, dy : Int) -> Point
    Point{x: @x + dx, y: @y + dy}  # OK: return new struct
  end
end

p = Point{x: 10, y: 20}
p.x = 5  # ERROR: cannot modify struct field
q = p.moved(5, 5)  # OK: creates new Point{x: 15, y: 25}
```

This immutability enables structs to be safely used as map keys and shared across goroutines.

### 12.5 Struct vs Class

| Feature | Struct | Class |
|---------|--------|-------|
| Semantics | Value type | Reference type |
| Equality | Value (all fields) | Identity (same object) |
| Mutability | Immutable (compile error) | Mutable |
| Inheritance | No | Yes |
| Constructor | Auto-generated | Manual (`initialize`) |
| Go output | `struct` (value) | `struct` (pointer) |

Use structs for: coordinates, configurations, DTOs, composite keys.
Use classes for: entities with identity, objects with behavior, mutable state.

---

## 13. Interfaces and Modules

### 13.1 Interface Definition

```ruby
interface Reader
  def read(buf : Bytes) -> (Int, error)
end

interface Writer
  def write(data : Bytes) -> (Int, error)
end

# Interface composition
interface ReadWriter < Reader, Writer
end

interface Closer
  def close -> error
end

interface ReadWriteCloser < Reader, Writer, Closer
end
```

### 13.2 Structural Typing

Classes satisfy interfaces automatically if they have matching methods:

```ruby
interface Speaker
  def speak -> String
end

class Dog
  def speak -> String
    "Woof!"
  end
end

class Cat
  def speak -> String
    "Meow!"
  end
end

def announce(s : Speaker)
  puts s.speak
end

announce(Dog.new)  # OK - Dog has speak method
announce(Cat.new)  # OK - Cat has speak method
```

### 13.3 Explicit Implementation

Use `implements` for compile-time verification:

```ruby
class FileReader implements Reader, Closer
  def read(buf : Bytes) -> (Int, error)
    # ...
  end

  def close -> error
    # ...
  end
end
```

If a class declares `implements` but lacks a method, compilation fails.

### 13.4 Type Checking

```ruby
obj.is_a?(Writer)              # returns Bool
w = obj.as(Writer)             # returns Writer?
w = obj.as(Writer) ?? fallback # with default

# In if-let
if let w = obj.as(Writer)
  w.write("hello")
end

# Type switch
case_type obj
when s : String
  puts "string: #{s}"
when w : Writer
  w.write("data")
else
  puts "unknown"
end
```

### 13.5 The `any` Type

`any` represents Go's empty interface `any` (or `interface{}`):

```ruby
def debug(value : any)
  p value
end

items : Array<any> = [1, "two", 3.0]
```

**Rules:**
- Any type can be assigned to `any`
- Must use type assertion (`as`) to recover concrete type
- Avoid `any` when possible—prefer generics or interfaces

**`any` vs generics:**
- Use generics (`<T>`) when all elements have the same type: `Array<T>`, `Map<K, V>`
- Use `any` only for truly heterogeneous collections or interop with untyped Go APIs
- `Array<any>` loses type safety; `Array<T>` preserves it

---

## 14. Modules (Mixins)

### 14.1 Module Definition

```ruby
module Loggable
  def log(msg : String)
    puts "[LOG] #{msg}"
  end

  def debug(msg : String)
    puts "[DEBUG] #{msg}"
  end
end

class Worker
  include Loggable

  def work
    log("Starting work")
    # ...
    log("Work complete")
  end
end

Worker.new.log("Hello")  # "[LOG] Hello"
```

### 14.2 Module State

Modules can define instance variables that become part of including classes:

```ruby
module Timestamped
  @created_at : Time?
  @updated_at : Time?

  def touch
    now = Time.now
    @created_at ||= now
    @updated_at = now
  end
end

class Document
  include Timestamped

  def save
    touch
    # ...
  end
end
```

### 14.3 Module Namespacing

Modules also serve as namespaces:

```ruby
module Http
  class Client
    def get(url : String) -> Response
      # ...
    end
  end

  class Response
    getter status : Int
    getter body : String
  end

  def self.get(url : String) -> Response
    Client.new.get(url)
  end
end

response = Http.get("https://example.com")
client = Http::Client.new
```

**Note:** Module composition is resolved entirely at compile time. Methods from included modules are copied into the including class.

### 14.4 Module Conflicts (Diamond Problem)

When multiple modules define the same method:

```ruby
module A
  def greet
    "Hello from A"
  end
end

module B
  def greet
    "Hello from B"
  end
end

class C
  include A
  include B  # B's greet wins (last included)
end

# Override to provide custom behavior
class D
  include A
  include B

  def greet
    "Hello from D"  # class method wins
  end
end
```

**Resolution rules:**
- Later includes override earlier ones
- Class methods override all included methods
- To use both, rename or wrap in separate methods

---

## 15. Visibility

### 15.1 Export with `pub`

`pub` exports to Go (uppercase). No `pub` = internal (lowercase).

```ruby
def helper(x)      # internal (lowercase in Go)
  x * 2
end

pub def double(x)  # exported (uppercase in Go)
  helper(x)
end

pub class Counter
  pub def inc
    @count += 1
  end

  def reset      # internal method
    @count = 0
  end
end
```

### 15.2 Class Visibility Levels

```ruby
class User
  # Public - accessible from anywhere (with pub)
  pub def name -> String
    @name
  end

  # Package-private - accessible within same package (default)
  def validate
    # ...
  end

  # Private - accessible only within this class
  private def encrypt_password
    # ...
  end
end
```

### 15.3 Visibility Summary

| Modifier | Go Output | Accessible From |
|----------|-----------|-----------------|
| `pub` | `Uppercase` | Anywhere (exported) |
| (none) | `lowercase` | Same package |
| `private` | `lowercase` | Same class only (compiler-enforced) |

### 15.4 Naming Transformations

Rugby snake_case transforms to Go camelCase/CamelCase:

| Rugby | Go (internal) | Go (pub) |
|-------|---------------|----------|
| `find_user` | `findUser` | `FindUser` |
| `user_id` | `userID` | `UserID` |
| `http_client` | `httpClient` | `HTTPClient` |
| `to_json` | `toJSON` | `ToJSON` |

Acronyms (`id`, `http`, `json`, `url`, `api`, `html`, `css`, `sql`) are uppercased.

---

## 16. Errors

Rugby follows Go's philosophy: **errors are values**, returned explicitly and handled locally.

### 16.1 Error Signatures

```ruby
def read_file(path : String) -> (Bytes, error)
  os.read_file(path)
end

def write_file(path : String, data : Bytes) -> error
  os.write_file(path, data)
end
```

**The `error` type:** Rugby uses lowercase `error` to match Go's `error` interface. Any type with an `error -> String` method satisfies this interface.

### 16.2 Creating Errors

```ruby
# Simple error with message
return nil, errors.new("something went wrong")

# Formatted error
return nil, fmt.errorf("invalid id: %d", id)

# Wrapping errors (preserves cause chain)
return nil, fmt.errorf("load config: %w", err)
```

### 16.3 Custom Error Types

```ruby
class ValidationError
  def initialize(@field : String, @message : String)
  end

  def error -> String
    "#{@field}: #{@message}"
  end
end

class NotFoundError
  def initialize(@resource : String, @id : any)
  end

  def error -> String
    "#{@resource} not found: #{@id}"
  end
end

def find_user(id : Int) -> (User, error)
  user = db.find(id)
  return nil, NotFoundError.new("User", id) if user.nil?
  user, nil
end
```

Any type with an `error -> String` method satisfies the `error` interface.

### 16.4 Postfix Bang Operator (`!`)

Propagates errors to the caller (like Rust's `?`):

```ruby
def load_config(path : String) -> (Config, error)
  data = os.read_file(path)!
  config = parse_json(data)!
  config, nil
end
```

**Rules:**
- Enclosing function must return `error`
- `(T, error)` -> `call!` evaluates to `T`
- `error` -> `call!` is control-flow only
- Binds tighter than all operators except member access/calls
- In `main` or scripts, prints error and exits (see 16.5)

**Chaining:**

```ruby
user = load_file(path)!.parse_json()!.get_user()!
```

### 16.5 The `rescue` Keyword

Inline error handling with fallback values:

```ruby
# Inline default
port = env.get("PORT").to_i() rescue 8080

# Block form
data = os.read_file(path) rescue do
  log.warn("using default config")
  default_config
end

# With error binding
data = os.read_file(path) rescue => err do
  log.warn("read failed: #{err}")
  return nil, fmt.errorf("load_config: %w", err)
end
```

**Rules:**
- Inline form: `default_expr` must match type `T`
- Block form: last expression becomes the value
- `rescue` and `!` are mutually exclusive
- For `error`-only returns, use block form (no value to replace)

### 16.6 Script Mode and `main`

In scripts and `def main`, `!` prints error to stderr and exits with code 1.

```ruby
data = os.read_file("config.json")!
puts data.length
```

### 16.7 Explicit Handling

Standard Go-style error handling:

```ruby
data, err = os.read_file(path)
if err != nil
  log.error("failed to read #{path}: #{err}")
  return nil, err
end
```

### 16.8 Error Wrapping

```ruby
def load_user(id : Int) -> (User, error)
  data, err = db.query(sql, id)
  if err != nil
    return nil, fmt.errorf("load_user(%d): %w", id, err)
  end
  parse_user(data)!
end
```

### 16.9 Error Utilities

```ruby
# Test error type (wrapping-aware)
if errors.is?(err, os.ErrNotExist)
  puts "file not found"
end

# Extract specific error type (returns Type?)
if let path_err = errors.as(err, os.PathError)
  puts "path error: #{path_err.path}"
end
```

### 16.10 Panics

For unrecoverable programmer errors:

```ruby
panic "unreachable"
panic "invalid state: #{state}"
```

Use `panic` only for invariants/bugs. Use `error` for expected failures.

### 16.11 Common Patterns

```ruby
# Retry with fallback
def fetch_config -> (Config, error)
  data = http.get(primary_url) rescue => err do
    log.warn("primary failed: #{err}")
    http.get(backup_url)!
  end
  parse_config(data)!
end

# Wrap and propagate
def load_user(id : Int) -> (User, error)
  row = db.query_row(sql, id) rescue => err do
    return nil, fmt.errorf("load_user(%d): %w", id, err)
  end
  parse_user(row)!
end

# Cleanup with defer
def process_file(path : String) -> error
  f = os.open(path)!
  defer f.close()

  process(f) rescue => err do
    log.error("processing failed: #{err}")
    return err
  end
  nil
end

# Ignore error, use default
timeout = config.get("timeout").to_i() rescue 30
```

---

## 17. Concurrency

### 17.1 Goroutines

```ruby
go fetch_url(url)

go do
  puts "background"
end
```

### 17.2 Channels

```ruby
ch = Chan<Int>.new(10)   # buffered
ch = Chan<Int>.new       # unbuffered

ch << val                # send
val = ch.receive         # blocking receive
val = ch.try_receive     # non-blocking, returns T?
ch.close

for msg in ch            # iterate until closed
  process(msg)
end
```

### 17.3 Select

```ruby
select
when val = ch1.receive
  puts "received #{val}"
when ch2 << 42
  puts "sent to ch2"
when timeout.receive
  puts "timed out"
else
  puts "no communication ready"
end
```

- Executes the first ready case (random if multiple ready)
- `else` makes select non-blocking
- Variables bound in `when` are scoped to that branch

### 17.4 Tasks (`spawn` and `await`)

```ruby
t = spawn { expensive_work() }  # Task<T>, runs immediately
result = await t                # blocks until complete

# With errors
t = spawn { os.read_file(path) }  # Task<(Bytes, error)>
data = await(t)!                  # parentheses required for !
```

- `spawn` returns `Task<T>` where `T` is the block's return type
- `await` blocks until completion; awaiting twice is a compile error
- Panics in tasks terminate the program (Go semantics)

### 17.5 Structured Concurrency (`concurrently`)

Scoped concurrency that ensures cleanup when the block exits:

```ruby
concurrently -> (scope) do
  a = scope.spawn { fetch_a() }
  b = scope.spawn { fetch_b() }

  x = await a
  y = await b
  x + y
end
```

**Rules:**
- `scope.spawn` creates tasks owned by the scope
- On block exit, the scope cancels `scope.ctx` and waits for all tasks
- Tasks cannot outlive their scope (compile error if task handle escapes)
- `concurrently` is an expression; returns the block's last value

**With errors:**

```ruby
result, err = concurrently -> (scope) do
  a = scope.spawn { fetch_a() }
  b = scope.spawn { fetch_b() }

  val_a = await(a)!
  val_b = await(b)!

  process(val_a, val_b)
end
```

Using `!` triggers early exit, cancelling sibling tasks.

**Detached tasks:** `spawn` (not `scope.spawn`) creates tasks that may outlive the function. Prefer `concurrently` for tasks that must complete safely.

### 17.6 Synchronization Primitives

```ruby
# Mutex
mutex = Mutex.new
mutex.lock
# critical section
mutex.unlock

# Or with lambda
mutex.synchronize -> do
  # critical section
end

# RWMutex
rw = RWMutex.new
rw.read_lock    # multiple readers OK
rw.read_unlock
rw.write_lock   # exclusive access
rw.write_unlock

# WaitGroup
wg = WaitGroup.new
wg.add(3)
3.times do
  go do
    # work
    wg.done
  end
end
wg.wait

# Once
once = Once.new
once.do do
  # executed exactly once
end
```

### 17.7 Choosing a Pattern

| Need | Use |
|------|-----|
| Fire-and-forget background work | `go do ... end` |
| Low-level channel communication | `Chan<T>`, `select` |
| Run work and get result | `spawn` + `await` |
| Multiple concurrent tasks with cleanup | `concurrently` |
| Cancellable, scoped work | `concurrently` with `scope.ctx` |
| Shared mutable state | `Mutex` or `RWMutex` |
| One-time initialization | `Once` |
| Waiting for multiple goroutines | `WaitGroup` |

### 17.8 Examples

**Simple parallel fetch:**
```ruby
def fetch_both(url1 : String, url2 : String) -> (String, String, error)
  concurrently -> (scope) do
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
def process_items(items : Array<Item>)
  work = Chan<Item>.new(100)
  done = Chan<Bool>.new

  # Start workers
  for _ in 0...4
    go do
      for item in work
        process(item)
      end
      done << true
    end
  end

  # Feed work
  for item in items
    work << item
  end
  work.close

  # Wait for workers
  for _ in 0...4
    done.receive
  end
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

### 17.9 Compilation Summary

| Rugby | Go |
|-------|-----|
| `go expr` | `go expr` |
| `go do ... end` | `go func() { ... }()` |
| `Chan<T>.new(n)` | `make(chan T, n)` |
| `spawn { expr }` | goroutine + channel for result |
| `await t` | receive from task's channel |
| `concurrently do \|s\| ... end` | `context.WithCancel` + `sync.WaitGroup` |
| `scope.spawn { ... }` | goroutine tracked by WaitGroup |
| `scope.ctx` | `context.Context` |
| `Mutex` | `sync.Mutex` |
| `RWMutex` | `sync.RWMutex` |
| `WaitGroup` | `sync.WaitGroup` |
| `Once` | `sync.Once` |

---

## 18. Go Interop

### 18.1 Imports

```ruby
import net/http
import encoding/json as json
import github.com/user/package

http.Get(url)              # call Go functions
io.read_all(r)             # snake_case -> ReadAll
defer resp.Body.Close()    # defer works as expected
```

### 18.2 Name Mapping

Rugby snake_case automatically maps to Go CamelCase:

| Rugby | Go |
|-------|-----|
| `http.get(url)` | `http.Get(url)` |
| `io.read_all(r)` | `io.ReadAll(r)` |
| `json.new_decoder(r)` | `json.NewDecoder(r)` |
| `os.read_file(path)` | `os.ReadFile(path)` |

**Explicit Go names:**

```ruby
# When auto-mapping doesn't work
resp = http.`Get`(url)  # use backticks for exact Go name
```

### 18.3 Type Mapping

| Rugby | Go |
|-------|-----|
| `Int` | `int` |
| `String` | `string` |
| `Array<T>` | `[]T` |
| `Map<K, V>` | `map[K]V` |
| `Bytes` | `[]byte` |
| `any` | `interface{}` |
| `error` | `error` |

### 18.4 Working with Go Types

```ruby
import net/http

def fetch(url : String) -> (String, error)
  resp, err = http.Get(url)
  return "", err if err != nil
  defer resp.Body.Close()

  body = io.read_all(resp.Body)!
  String(body), nil
end
```

### 18.5 Defer

```ruby
def process_file(path : String) -> error
  f = os.open(path)!
  defer f.close()

  # f.close() called when function returns
  process(f)
end

# Multiple defers execute LIFO
def example
  defer puts("first")
  defer puts("second")
  defer puts("third")
  # Output: third, second, first
end
```

### 18.6 Embedding Go Types

```ruby
class MyReader
  embed io.Reader  # embed Go interface

  def initialize(@inner : io.Reader)
  end

  def read(p : Bytes) -> (Int, error)
    @inner.read(p)
  end
end
```

---

## 19. Runtime Package

The `rugby/runtime` package provides Ruby-like methods for Go types.

### 19.1 Array Methods (`Array<T>`)

**Iteration:**
* `each -> (x) { }` -> `for _, x := range arr { ... }`
* `each_with_index -> (x, i) { }` -> `for i, x := range arr { ... }`

**Transformation:**
* `map -> (x) { }` -> `runtime.Map(arr, fn)` -> `[]R`
* `select -> (x) { }` / `filter -> (x) { }` -> `runtime.Select(arr, fn)` -> `[]T`
* `reject -> (x) { }` -> `runtime.Reject(arr, fn)` -> `[]T`
* `compact` -> `runtime.Compact(arr)` -> filters nil/zero values
* `flatten` -> `runtime.Flatten(arr)` -> flattens nested arrays
* `uniq` -> `runtime.Uniq(arr)` -> removes duplicates

**Search:**
* `find -> (x) { }` / `detect -> (x) { }` -> `runtime.Find(arr, fn)` -> `T?`
* `find_index -> (x) { }` -> `runtime.FindIndex(arr, fn)` -> `Int?`
* `any? -> (x) { }` -> `runtime.Any(arr, fn)` -> `Bool`
* `all? -> (x) { }` -> `runtime.All(arr, fn)` -> `Bool`
* `none? -> (x) { }` -> `runtime.None(arr, fn)` -> `Bool`
* `include?(val)` / `contains?(val)` -> `runtime.Contains(arr, val)` -> `Bool`
* `index(val)` -> `runtime.Index(arr, val)` -> `Int?`

**Aggregation:**
* `reduce(init) -> (acc, x) { }` -> `runtime.Reduce(arr, init, fn)` -> `R`
* `sum` -> `runtime.Sum(arr)` (numeric arrays)
* `min` / `max` -> `runtime.Min(arr)` / `runtime.Max(arr)` -> `T?`
* `min_by -> (x) { }` / `max_by -> (x) { }` -> `runtime.MinBy` / `runtime.MaxBy` -> `T?`

**Access:**
* `first` -> `runtime.First(arr)` -> `T?`
* `last` -> `runtime.Last(arr)` -> `T?`
* `take(n)` -> `runtime.Take(arr, n)` -> `[]T`
* `drop(n)` -> `runtime.Drop(arr, n)` -> `[]T`
* `length` / `size` -> `len(arr)` (inlined)
* `empty?` -> `len(arr) == 0` (inlined)

**Append operator (`<<`):**
```ruby
arr << item           # append single item
arr << a << b << c    # chainable (returns new slice)
arr = arr << item     # reassign to capture result
```

**Mutation (in-place):**
* `push(item)` / `append(item)` -> modifies in place
* `pop` -> removes and returns last element
* `shift` -> removes and returns first element
* `unshift(item)` -> prepends item
* `reverse!` -> `runtime.ReverseInPlace(arr)`
* `sort!` -> `runtime.SortInPlace(arr)`
* `sort_by! -> (x) { }` -> `runtime.SortByInPlace(arr, fn)`

**Non-mutating (returns copy):**
* `reversed` -> `runtime.Reversed(arr)` -> new slice
* `sorted` -> `runtime.Sorted(arr)` -> new slice
* `sorted_by -> (x) { }` -> `runtime.SortedBy(arr, fn)` -> new slice

### 19.2 Map Methods (`Map<K, V>`)

**Iteration:**
* `each -> (k, v) { }` -> `runtime.MapEach(m, fn)`
* `each_key -> (k) { }` -> `runtime.MapEachKey(m, fn)`
* `each_value -> (v) { }` -> `runtime.MapEachValue(m, fn)`

**Access:**
* `keys` -> `runtime.Keys(m)` -> `[]K`
* `values` -> `runtime.Values(m)` -> `[]V`
* `length` / `size` -> `len(m)` (inlined)
* `empty?` -> `len(m) == 0` (inlined)
* `has_key?(k)` / `key?(k)` -> `_, ok := m[k]` (inlined)
* `fetch(k, default)` -> `runtime.Fetch(m, k, default)` -> `V`
* `get(k)` -> `runtime.Get(m, k)` -> `V?`

**Transformation:**
* `select -> (k, v) { }` -> `runtime.MapSelect(m, fn)` -> `Map<K, V>`
* `reject -> (k, v) { }` -> `runtime.MapReject(m, fn)` -> `Map<K, V>`
* `map_values -> (v) { }` -> `runtime.MapValues(m, fn)` -> `Map<K, R>`
* `merge(other)` -> `runtime.Merge(m, other)` -> `Map<K, V>`

**Mutation:**
* `delete(k)` -> removes key
* `clear` -> removes all keys

### 19.3 String Methods

**Query:**
* `length` / `size` -> `len(s)` (inlined, byte length)
* `char_length` -> `runtime.CharLength(s)` (rune count)
* `empty?` -> `s == ""` (inlined)
* `include?(sub)` / `contains?(sub)` -> `strings.Contains` (inlined)
* `start_with?(prefix)` -> `strings.HasPrefix` (inlined)
* `end_with?(suffix)` -> `strings.HasSuffix` (inlined)
* `index(sub)` -> `strings.Index` -> `Int?`

**Transformation:**
* `upcase` -> `strings.ToUpper` (inlined)
* `downcase` -> `strings.ToLower` (inlined)
* `capitalize` -> first char upper, rest lower
* `strip` / `trim` -> `strings.TrimSpace` (inlined)
* `lstrip` / `rstrip` -> `strings.TrimLeft` / `TrimRight`
* `replace(old, new)` -> `strings.ReplaceAll` (inlined)
* `reverse` -> `runtime.StringReverse(s)`
* `chomp` -> remove trailing newline
* `center(width)` / `ljust(width)` / `rjust(width)` -> padding

**Splitting/Joining:**
* `split(sep)` -> `strings.Split` (inlined)
* `split(re)` -> `re.Split(s, -1)` for regex
* `chars` -> `runtime.Chars(s)` -> `[]String`
* `lines` -> `strings.Split(s, "\n")` (inlined)
* `bytes` -> `[]byte(s)` (inlined)

**Conversion:**
* `to_i` -> `runtime.StringToInt(s)` -> `(Int, error)`
* `to_f` -> `runtime.StringToFloat(s)` -> `(Float, error)`
* `to_sym` -> Symbol (returns self, type conversion)

### 19.4 Integer Methods

**Predicates:**
* `even?` -> `n % 2 == 0` (inlined)
* `odd?` -> `n % 2 != 0` (inlined)
* `zero?` -> `n == 0` (inlined)
* `positive?` -> `n > 0` (inlined)
* `negative?` -> `n < 0` (inlined)

**Math:**
* `abs` -> `runtime.Abs(n)`
* `clamp(min, max)` -> `runtime.Clamp(n, min, max)`
* `**` (power) -> `runtime.Pow(base, exp)` or `math.Pow`

**Iteration:**
* `times -> (i) { }` -> `runtime.Times(n, fn)`
* `upto(max) -> (i) { }` -> `runtime.Upto(n, max, fn)`
* `downto(min) -> (i) { }` -> `runtime.Downto(n, min, fn)`

**Conversion:**
* `to_s` -> `strconv.Itoa(n)` (inlined)
* `to_f` -> `float64(n)` (inlined)

### 19.5 Float Methods

**Rounding:**
* `floor` -> `math.Floor` (inlined)
* `ceil` -> `math.Ceil` (inlined)
* `round` -> `math.Round` (inlined)
* `round(places)` -> round to decimal places
* `truncate` -> `math.Trunc` (inlined)

**Predicates:**
* `zero?` -> `f == 0.0` (inlined)
* `positive?` / `negative?` -> comparison (inlined)
* `nan?` -> `math.IsNaN(f)` (inlined)
* `infinite?` -> `math.IsInf(f, 0)` (inlined)

**Conversion:**
* `to_i` -> `int(f)` (inlined)
* `to_s` -> `strconv.FormatFloat` (inlined)

### 19.6 Global Functions (Kernel)

```ruby
puts "message"           # print with newline
print "message"          # print without newline
p value                  # inspect/debug print
gets                     # read line from stdin -> String?
exit(code)               # exit with status code
sleep(seconds)           # sleep for duration
rand                     # random float 0.0..1.0
rand(max)                # random int 0..max-1
rand(range)              # random in range
```

---

## 20. Testing

### 20.1 Test Functions

```ruby
# user_test.rg

def test_user_creation
  user = User.new("Alice", 30)
  assert user.name == "Alice"
  assert user.age == 30
end

def test_user_greeting
  user = User.new("Bob", 25)
  assert_eq user.greet, "Hello, Bob!"
end

def test_validation_failure
  assert_raises ValidationError do
    User.new("", -1)
  end
end
```

### 20.2 Assertions

```ruby
assert condition                      # fails if false
assert condition, "custom message"    # with message
assert_eq actual, expected            # equality
assert_ne actual, expected            # inequality
assert_nil value                      # must be nil
assert_not_nil value                  # must not be nil
assert_raises ErrorType do ... end    # must raise error
assert_match /pattern/, string        # regex match
```

### 20.3 Test Organization

```ruby
# Group related tests
describe "User" do
  describe "creation" do
    def test_with_valid_args
      # ...
    end

    def test_with_invalid_args
      # ...
    end
  end

  describe "authentication" do
    def setup
      @user = User.new("Alice", 30)
    end

    def teardown
      # cleanup
    end

    def test_login
      # ...
    end
  end
end
```

### 20.4 Running Tests

```bash
rugby test                    # run all tests
rugby test user_test.rg       # run specific file
rugby test -v                 # verbose output
rugby test -f "pattern"       # filter by name
```

---

## 21. Diagnostics

### 21.1 Compile-Time Errors

* Type errors and unresolved identifiers are compile-time errors
* Clear source locations with line and column numbers
* Suggestions for common mistakes

### 21.2 Error Messages

Good error messages include:
* Source location (file:line:column)
* What went wrong
* Why it's wrong
* Suggestions to fix

```
user.rg:15:10: error: type mismatch
  expected: String
  got: Int

  name = 42
         ^^

  suggestion: use `42.to_s` to convert Int to String
```

### 21.3 Go Interop Diagnostics

When Go interop fails:
* Show Rugby source span
* Show intended Go package/type
* Suggest candidates (e.g., `read_all` -> `ReadAll`)

```
fetch.rg:8:12: error: unknown method `read_all` on `io`

  data = io.read_all(reader)
            ^^^^^^^^

  did you mean: io.ReadAll (note: Go uses CamelCase)
```

---

## 22. Differences from Ruby

### 22.1 Static Typing

```ruby
# Ruby
def add(a, b)
  a + b
end
add(1, 2)      # 3
add("a", "b")  # "ab"

# Rugby - parameters must be typed
def add(a : Int, b : Int) -> Int
  a + b
end
add(1, 2)      # 3
add("a", "b")  # Compile error
```

### 22.2 No Metaprogramming

```ruby
# Ruby
class User
  attr_accessor :name
  define_method(:greet) { "Hello" }
end
User.class_eval { ... }

# Rugby - no runtime modification
class User
  property name : String  # explicit accessor

  def greet
    "Hello"
  end
end
```

### 22.3 No Implicit Truthiness

```ruby
# Ruby
if user        # truthy if not nil/false
if items.any?  # truthy if not empty

# Rugby - must be Bool
if user.ok?          # explicit nil check
if items.any?        # OK, returns Bool
if items.length > 0  # explicit comparison
```

### 22.4 Lambda Control Flow

```ruby
# Ruby - return/break/next work in blocks
def find_user
  users.each { |u| return u if u.active? }
  nil
end

# Rugby - lambdas return from the lambda only
def find_user
  users.find -> (u) { u.active? }  # use find instead
end

# For control flow affecting the enclosing function, use for loops
def find_admin
  for user in users
    return user if user.admin?
  end
  nil
end
```

In Rugby, `return` inside a lambda returns from the lambda, not the enclosing function. Use `for` loops when you need `return`, `break`, or `next` to affect the outer scope.

### 22.5 No Open Classes

```ruby
# Ruby
class String
  def shout
    self.upcase + "!"
  end
end

# Rugby - use extension methods (future) or wrapper
module StringExt
  def shout(s : String) -> String
    s.upcase + "!"
  end
end
```

### 22.6 Different Iteration

```ruby
# Ruby - many ways
for i in 0..5 do ... end
(0..5).each { |i| ... }
5.times { |i| ... }

# Rugby - for loops preferred for side effects
for i in 0..5
  puts i
end

# Lambdas for transformation
items.map -> (i) { i * 2 }
```

---

## 23. Differences from Go

### 23.1 Ruby-Style Syntax

```go
// Go
func add(a, b int) int {
    return a + b
}

// Rugby
def add(a : Int, b : Int) -> Int
  a + b
end
```

### 23.2 Implicit Returns

```go
// Go - explicit return required
func double(x int) int {
    return x * 2
}

// Rugby - last expression is return value
def double(x : Int) -> Int
  x * 2
end
```

### 23.3 Optional Types vs Pointers

```go
// Go - use pointers for optional
func find(id int) *User {
    return nil
}
if user := find(1); user != nil { ... }

// Rugby - explicit optional type
def find(id : Int) -> User?
  nil
end
if let user = find(1)
  # ...
end
```

### 23.4 Error Handling Syntax

```go
// Go
data, err := os.ReadFile(path)
if err != nil {
    return nil, err
}

// Rugby - same semantics, less boilerplate
data = os.read_file(path)!
```

### 23.5 Iterators

```go
// Go
for i, item := range items {
    fmt.Println(item)
}

// Rugby
for item in items
  puts item
end

# Or with index
items.each_with_index -> (item, i) { puts "#{i}: #{item}" }
```

### 23.6 Method Chaining

```go
// Go - awkward nesting
filtered := Filter(items, predicate)
mapped := Map(filtered, transform)
result := Reduce(mapped, initial, accumulate)

// Rugby - natural chaining
result = items.select -> (x) { predicate(x) }
              .map -> (x) { transform(x) }
              .reduce(initial) -> (acc, x) { accumulate(acc, x) }
```

---

## 24. Appendix: Grammar Summary

This grammar is illustrative, showing the major language constructs. Some productions (e.g., `defer`, `go`, `spawn`, `select`, operators) are omitted for brevity. See the prose sections for complete details.

```ebnf
program        = { import | declaration | statement } ;

import         = "import" package_path [ "as" ident ] ;

declaration    = const_decl | type_decl | func_decl | class_decl
               | struct_decl | interface_decl | module_decl | enum_decl ;

const_decl     = "const" IDENT [ ":" type ] "=" expr ;
type_decl      = "type" TYPE_NAME "=" type ;

func_decl      = [ "pub" ] "def" [ "self." ] ident [ type_params ]
                 "(" [ params ] ")" [ "->" type ] { statement } "end" ;

class_decl     = [ "pub" ] "class" TYPE_NAME
                 [ type_params ] [ "<" TYPE_NAME ]
                 [ "implements" TYPE_NAME { "," TYPE_NAME } ]
                 class_body "end" ;

struct_decl    = [ "pub" ] "struct" TYPE_NAME [ type_params ]
                 struct_body "end" ;

interface_decl = [ "pub" ] "interface" TYPE_NAME [ type_params ]
                 [ "<" TYPE_NAME { "," TYPE_NAME } ]
                 interface_body "end" ;

module_decl    = [ "pub" ] "module" TYPE_NAME module_body "end" ;

enum_decl      = [ "pub" ] "enum" TYPE_NAME [ type_params ]
                 enum_body "end" ;

type           = TYPE_NAME [ type_args ] [ "?" ]
               | "(" type { "," type } ")"
               | "(" [ type { "," type } ] ")" "->" type ;

type_params    = "<" type_param { "," type_param } ">" ;
type_param     = TYPE_NAME [ ":" type { "&" type } ] ;
type_args      = "<" type { "," type } ">" ;

statement      = if_stmt | unless_stmt | case_stmt | case_type_stmt
               | for_stmt | while_stmt | until_stmt | loop_stmt
               | return_stmt | break_stmt | next_stmt
               | assign_stmt | expr_stmt ;

expr           = literal | ident | instance_var | class_var
               | unary_expr | binary_expr | call_expr | index_expr
               | member_expr | lambda_expr
               | if_expr | case_expr | case_type_expr | "(" expr ")" ;

lambda_body    = "do" { statement } "end"
               | "{" expr "}" ;

lambda_expr    = "->" [ "(" [ params ] ")" ] [ "->" type ] lambda_body ;
```

---

## 25. Appendix: Compilation Examples

### Simple Function

```ruby
# Rugby
def greet(name : String) -> String
  "Hello, #{name}!"
end
```

```go
// Go
func greet(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}
```

### Class with Methods

```ruby
# Rugby
class Counter
  def initialize(@count : Int = 0)
  end

  def inc
    @count += 1
  end

  def value -> Int
    @count
  end
end
```

```go
// Go
type Counter struct {
    count int
}

func NewCounter(count int) *Counter {
    return &Counter{count: count}
}

func (c *Counter) Inc() {
    c.count++
}

func (c *Counter) Value() int {
    return c.count
}
```

### Error Handling

```ruby
# Rugby
def load_config(path : String) -> (Config, error)
  data = os.read_file(path)!
  config = parse_json(data)!
  config, nil
end
```

```go
// Go
func loadConfig(path string) (Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return Config{}, err
    }
    config, err := parseJSON(data)
    if err != nil {
        return Config{}, err
    }
    return config, nil
}
```

### Lambdas and Iteration

```ruby
# Rugby
names = users.select -> (u) { u.active? }
             .map -> (u) { u.name }
             .sorted
```

```go
// Go
var filtered []*User
for _, u := range users {
    if u.Active() {
        filtered = append(filtered, u)
    }
}
var names []string
for _, u := range filtered {
    names = append(names, u.Name())
}
sort.Strings(names)
```

---

*Rugby: Write Ruby, run Go.*
