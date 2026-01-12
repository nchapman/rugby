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

### 3.4 Heredocs

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

Symbols (`:ok`, `:status`, `:not_found`) compile to Go strings.

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

#### Array Literals

```ruby
nums = [1, 2, 3]
words = %w{foo bar baz}        # ["foo", "bar", "baz"]
words = %W{hello #{name}}      # interpolated: ["hello", "world"]
all = [1, *rest, 4]            # splat: [1, 2, 3, 4]
```

Word arrays support any delimiter: `%w{...}`, `%w(...)`, `%w[...]`, `%w<...>`, `%w|...|`

#### Indexing

Negative indices count from the end:

```ruby
arr[-1]   # last element
arr[-2]   # second-to-last
str[-1]   # last character (Unicode-aware)
```

Out-of-bounds indices panic. For safe access, use `first`/`last` which return optionals.

#### Range Slicing

```ruby
arr[1..3]   # [1, 2, 3] (inclusive)
arr[1...3]  # [1, 2] (exclusive end)
arr[-3..-1] # last 3 elements
str[1..3]   # "ell"
```

Out-of-bounds ranges return empty. Negative indices count from end.

#### Map Literals

```ruby
m = {"name" => "Alice", "age" => 30}  # hash rocket
m = {name: "Alice", age: 30}          # symbol shorthand (same result)
m = {name:, age:}                     # implicit value (uses variables)
m = {**defaults, **overrides}         # double splat merges maps
```

**Note:** Map literals require parentheses with command syntax: `foo({a: 1})` not `foo {a: 1}` (parsed as block).

### 4.2.1 Range Type

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

### 4.2.2 Channel Type

```ruby
ch = Chan[Int].new(10)  # buffered
ch = Chan[Int].new      # unbuffered
```

See Section 13.2 for channel operations.

### 4.2.3 Task Type

```ruby
t = spawn { expensive_work() }  # Task[Result]
result = await t
```

See Section 13.4 for spawning and awaiting.

### 4.3 Type Inference

```ruby
x = 5           # x is Int (inferred)
y : Int64 = 5   # explicit type annotation

def add(a : Int, b : Int)  # parameters must be typed
def log(msg : any)         # use `any` for dynamic types
```

**Instance Variables:** Types resolved from:
1.  **Explicit Declarations:** `@field : Type` at the class level.
2.  **Initialize Inference:** Assignments within the `initialize` method.
3.  **Parameter Promotion:** `def initialize(@field : Type)` syntax.

```ruby
class User
  @role : String                # explicit declaration

  def initialize(@name : String) # parameter promotion
    @age = 0                     # inferred from initialize
  end
end
```

New instance variables may only be introduced in `initialize`. Non-optional fields must be assigned on all control-flow paths.

### 4.4 Optionals (`T?`)

`T?` represents optional values. Compiles to `(T, bool)` for returns, `runtime.Option[T]` for storage.

#### 4.4.1 Optional Operators

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

#### 4.4.2 Optional Methods

| Method | Description |
|--------|-------------|
| `ok?` / `present?` | Is value present? |
| `nil?` / `absent?` | Is value absent? |
| `unwrap` | Unwrap, panic if absent |
| `map { \|x\| }` | Transform if present → `R?` |
| `each { \|x\| }` | Execute block if present |

#### 4.4.3 Unwrapping Patterns

```ruby
# if let - scoped binding (preferred)
if let user = find_user(id)
  puts user.name
end

# Tuple unpacking - Go's comma-ok idiom
user, ok = find_user(id)
if ok
  puts user.name
end
```

| Pattern | Use case |
|---------|----------|
| `expr ?? default` | Fallback value |
| `expr&.method` | Safe navigation |
| `if let u = expr` | Conditional with binding |
| `val, ok = expr` | Need boolean separately |
| `expr.unwrap` | Absence is a bug (panics) |

### 4.5 The `nil` Keyword

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

## 5. Variables & Control Flow

### 5.1 Variables & Operators

```ruby
x = expr       # declares (new) or assigns (existing)
x += 5         # compound: +=, -=, *=, /=
x ||= default  # assign if nil/false/absent
```

### 5.2 Conditionals

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
else
  puts "error"
end

# case_type - type matching
case_type x
when String
  puts "string"
when Int
  puts "int"
end

# ternary (right-associative)
max = a > b ? a : b
```

### 5.3 Loops

```ruby
for item in items        # iterate collection
  return item if item.id == 5
end

for i in 0..10           # iterate range
  puts i
end

while running
  process_next()
end

until queue.empty?       # while false (inverse)
  handle(queue.pop)
end
```

Control: `break` (exit), `next` (continue), `return` (from function)

### 5.4 Blocks

Blocks are anonymous functions for data transformation:

```ruby
names = users.map { |u| u.name }
names = users.map do |u|
  u.name
end

# Symbol-to-proc shorthand
names.map(&:upcase)       # same as: names.map { |x| x.upcase }
users.select(&:active?)
```

- `do...end` for multi-line, `{...}` for single-line
- Last expression is the block's return value
- `return value` in a block returns `value` from the block (not the enclosing function - differs from Ruby)

### 5.5 Statement Modifiers

```ruby
return nil if id < 0
puts "error" unless valid?
break if done
```

### 5.6 Loop Modifiers

```ruby
puts items.shift while items.any?
process(queue.pop) until queue.empty?
```

Unlike statement modifiers (`if`/`unless`), loop modifiers execute repeatedly.

---

## 6. Functions

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

def parse(s) -> (Int, Bool)      # multiple returns
  return 0, false if s.empty?
  s.to_i, true
end

# Destructuring
n, ok = parse("42")     # both values
n, _ = parse("42")      # ignore second with _
```

### 6.4 Errors

```ruby
def read(path : String) -> (Bytes, error)
  os.read_file(path)
end
```

See Section 15 for `!` and `rescue` error handling.

### 6.5 Calling Convention

Parentheses are optional (command syntax):

```ruby
puts "hello"       # puts("hello")
add 1, 2           # add(1, 2)
foo.bar baz        # foo.bar(baz)
```

**Key rules:**
- Arguments end at newline, `do`, `{`, or `and`/`or`
- `foo -1` → `foo(-1)`, `foo - 1` → binary subtraction
- `foo [1]` → array arg, `foo[1]` → indexing
- Maps need parens: `foo({a: 1})` (braces start blocks)

---

## 7. Classes

```ruby
class User implements Identifiable
  @email : String               # explicit field
  property name : String        # declares field + getter/setter

  def initialize(@name : String, @age : Int)
    @email = "unknown"          # inferred field
  end
end

u = User.new("Alice", 30)       # auto-generated
```

**Instance variables** introduced only via:
1. Explicit declaration: `@field : Type`
2. Parameter promotion: `def initialize(@field : Type)`
3. First assignment in `initialize`

Methods use pointer receivers. All non-optional fields must be assigned in `initialize`. Within methods, `self` refers to the receiver (optional when calling other methods on same object).

### 7.4 Accessors

```ruby
getter age : Int        # declares @age + generates def age -> Int
setter name : String    # declares @name + generates def name=(v)
property email : String # both getter and setter
```

**Field access:** Inside class methods, use `@field`. Outside, access via methods only - direct field access is not exposed. Use `pub getter`/`pub property` to export accessors.

### 7.5 Inheritance

```ruby
class Child < Parent
  def name
    "Child"
  end
end

Child.new.hello  # calls Parent#hello with Child's method dispatch
```

- Single inheritance only (`class Child < Parent`)
- `super` calls parent's implementation
- Concrete types are invariant (use interfaces for polymorphism)

### 7.6 Interfaces

```ruby
interface Speaker
  def speak -> String
end

class User implements Speaker  # explicit (optional, compile-time check)
  def speak
    "Hello"
  end
end
```

Rugby uses structural typing - classes satisfy interfaces automatically if they have matching methods.

---

## 8. Modules (Mixins)

```ruby
module Loggable
  def log(msg)
    puts msg
  end
end

class Worker
  include Loggable
end

Worker.new.log("working")  # methods mixed in
```

---

## 9. Interfaces

```ruby
interface Reader
  def read -> String
end

interface Writer
  def write(data : String)
end

interface IO < Reader, Writer  # interface composition
  def close
end
```

### 9.2 Type Checking

```ruby
obj.is_a?(Writer)              # returns Bool
w = obj.as(Writer)             # returns Writer?
w = obj.as(Writer) ?? fallback # with default
```

### 9.3 The `any` Type

`any` represents Go's empty interface `interface{}`.

---

## 10. Visibility & Naming

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
end
```

**Naming:**
- Types: `CamelCase`, functions/methods: `snake_case`
- `method?` must return `Bool`
- `snake_case` → Go `camelCase`/`CamelCase` with acronym handling (`user_id` → `userID`)

---

## 11. Go Interop

```ruby
import net/http
import encoding/json as json

http.Get(url)              # call Go functions
io.read_all(r)             # snake_case → ReadAll
defer resp.Body.Close()    # defer works as expected
```

---

## 12. Runtime Package

The `rugby/runtime` package provides Ruby-like methods for Go types.

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

**Append operator (`<<`):**
```ruby
arr << item           # append single item
arr << a << b << c    # chainable (returns new slice)
arr = arr << item     # reassign to capture result
```

The `<<` operator works for both arrays and channels:
* Arrays: `runtime.ShiftLeft(arr, item)` → appends and returns new slice
* Channels: `runtime.ShiftLeft(ch, val)` → sends value and returns channel

**Note:** Unlike Ruby's in-place mutation, `<<` returns a new slice in Rugby. To update the original variable, reassign: `arr = arr << item`.

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

### 12.8 Global functions (kernel)

Top-level functions: `puts`, `print`, `p`, `gets`, `exit`, `sleep`, `rand`

---

## 13. Concurrency

### 13.1 Goroutines

```ruby
go fetch_url(url)

go do
  puts "background"
end
```

### 13.2 Channels

```ruby
ch = Chan[Int].new(10)   # buffered
ch << val                # send
val = ch.receive         # blocking receive
val = ch.try_receive     # non-blocking, returns T?
ch.close

for msg in ch            # iterate until closed
  process(msg)
end
```

### 13.3 Select

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

- Executes the first ready case (random if multiple ready)
- `else` makes select non-blocking
- Variables bound in `when` are scoped to that branch

### 13.4 Tasks (`spawn` and `await`)

```ruby
t = spawn { expensive_work() }  # Task[T], runs immediately
result = await t                # blocks until complete

# With errors
t = spawn { os.read_file(path) }  # Task[(Bytes, error)]
data = await(t)!                  # parentheses required for !
```

- `spawn` returns `Task[T]` where `T` is the block's return type
- `await` blocks until completion; awaiting twice is a compile error
- Panics in tasks terminate the program (Go semantics)

### 13.5 Structured Concurrency (`concurrently`)

Scoped concurrency that ensures cleanup when the block exits:

```ruby
concurrently do |scope|
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
result, err = concurrently do |scope|
  a = scope.spawn { fetch_a() }
  b = scope.spawn { fetch_b() }

  val_a = await(a)!
  val_b = await(b)!

  process(val_a, val_b)
end
```

Using `!` triggers early exit, cancelling sibling tasks.

**Detached tasks:** `spawn` (not `scope.spawn`) creates tasks that may outlive the function. Prefer `concurrently` for tasks that must complete safely.

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

Rugby follows Go's philosophy: **errors are values**, returned explicitly and handled locally.

### 15.1 Error Signatures

```ruby
def read_file(path : String) -> (Bytes, error)
  os.read_file(path)
end

def write_file(path : String, data : Bytes) -> error
  os.write_file(path, data)
end
```

### 15.2 Postfix Bang Operator (`!`)

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
- `(T, error)` → `call!` evaluates to `T`
- `error` → `call!` is control-flow only
- Binds tighter than all operators except member access/calls
- In `main` or scripts, prints error and exits (see 15.5)

**Chaining:**

```ruby
user = load_file(path)!.parse_json()!.get_user()!
```

### 15.3 The `rescue` Keyword

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

### 15.4 Script Mode and `main`

In scripts and `def main`, `!` prints error to stderr and exits with code 1.

```ruby
data = os.read_file("config.json")!
puts data.length
```

### 15.5 Explicit Handling

Standard Go-style error handling:

```ruby
data, err = os.read_file(path)
if err != nil
  log.error("failed to read #{path}: #{err}")
  return nil, err
end
```

### 15.6 Panics

For unrecoverable programmer errors:

```ruby
panic "unreachable"
panic "invalid state: #{state}"
```

Use `panic` only for invariants/bugs. Use `error` for expected failures.

### 15.7 Error Utilities

```ruby
# Test error type (wrapping-aware)
if error_is?(err, os.ErrNotExist)
  puts "file not found"
end

# Extract specific error type (returns Type?)
if let path_err = error_as(err, os.PathError)
  puts "path error: #{path_err.path}"
end
```

### 15.8 Common Patterns

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

