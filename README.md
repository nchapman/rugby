# Rugby

Rugby is a statically-typed language with Ruby's elegance that compiles to Go. Write expressive, readable code and get fast, standalone binaries with full access to Go's ecosystem.

```ruby
import "rugby/http"

resp = http.get("https://api.github.com/users/octocat")!
user = resp.json!
puts "#{user["name"]} has #{user["public_repos"]} repos"
```

## Installation

```bash
go install github.com/nchapman/rugby@latest
```

Verify it worked:

```bash
rugby --version
```

### From source

```bash
git clone https://github.com/nchapman/rugby.git
cd rugby
go install .
```

## Quick Start

### Hello World

```ruby
# hello.rg
puts "Hello, world!"
```

```bash
rugby run hello.rg
```

### Build a binary

```bash
rugby build hello.rg -o hello
./hello
```

### Interactive REPL

```bash
rugby repl
```

## Language Guide

### Variables and Types

Variables are declared on first assignment. Types are inferred but can be explicit. All type names use `PascalCase`.

```ruby
name = "Alice"              # String
age = 30                    # Int
price = 19.99               # Float
active = true               # Bool
tags = ["ruby", "go"]       # Array<String>
scores = {"alice" => 100}   # Map<String, Int>
status = :ok                # Symbol (compiles to string constant)
```

Explicit type annotation:

```ruby
count : Int = 0
ratio : Float = 0.5
items : Array<String> = []
```

### Functions

Functions use `def` and `end`. The last expression is returned implicitly.

```ruby
def greet(name : String) -> String
  "Hello, #{name}!"
end

def add(a : Int, b : Int) -> Int
  a + b
end

puts greet("World")  # Hello, World!
puts add(2, 3)       # 5
```

Multiple return values follow Go conventions:

```ruby
def divide(a : Int, b : Int) -> (Int, Error)
  return 0, errors.new("division by zero") if b == 0
  a / b, nil
end
```

### Control Flow

**If/elsif/else:**

```ruby
if score >= 90
  puts "A"
elsif score >= 80
  puts "B"
else
  puts "C"
end
```

**Unless** (inverse of if):

```ruby
unless user.nil?
  puts user.name
end
```

**Statement modifiers:**

```ruby
puts "adult" if age >= 18
return nil if id < 0
break if done
```

**For loops:**

```ruby
for item in items
  puts item
end

for i in 0..10
  puts i
end

for key, value in map
  puts "#{key}: #{value}"
end
```

**While loops:**

```ruby
while running
  process_next()
end
```

**Case expressions:**

```ruby
case status
when 200, 201
  puts "success"
when 400..499
  puts "client error"
when 500..599
  puts "server error"
else
  puts "unknown"
end
```

### Lambdas and Iteration

Lambdas are anonymous functions using `->` syntax. Use braces `{ }` for single-line, `do...end` for multi-line:

```ruby
# Single-line lambdas
squares = nums.map -> (n) { n * n }
evens = nums.select -> (n) { n.even? }

# Multi-line lambdas
results = items.map -> (item) do
  processed = transform(item)
  validate(processed)
end

# Iteration
[1, 2, 3].each -> (n) { puts n }

# Finding
first_big = items.find -> (x) { x.size > 100 }

# Aggregation
sum = [1, 2, 3].reduce(0) -> (acc, n) { acc + n }  # 6

# Integer iteration
5.times -> (i) { puts i }
1.upto(10) -> (i) { puts i }
```

**Symbol-to-proc** for concise method calls:

```ruby
names = users.map(&:name)           # same as: users.map -> (u) { u.name }
active = users.select(&:active?)    # same as: users.select -> (u) { u.active? }
lengths = strings.map(&:length)
```

**Calling lambdas:**

```ruby
double = -> (x) { x * 2 }
result = double.(5)      # 10 (preferred)
result = double.call(5)  # 10 (alternative)
```

**Lambdas vs for loops:** Use lambdas for transformation. Use `for` loops when you need `break`, `next`, or `return` from the enclosing function:

```ruby
# Lambda - return exits the lambda only
users.each -> (u) do
  return if u.inactive?  # skips this user, continues iteration
  process(u)
end

# For loop - full control flow
def find_admin(users : Array<User>) -> User?
  for user in users
    return user if user.admin?  # returns from function
  end
  nil
end
```

### Classes

Classes define structs with methods. Instance variables use `@`.

```ruby
class User
  property name : String
  property email : String
  getter age : Int

  def initialize(@name : String, @email : String, @age : Int)
  end

  def adult? -> Bool
    @age >= 18
  end

  def greeting -> String
    "Hi, I'm #{@name}"
  end
end

user = User.new("Alice", "alice@example.com", 30)
puts user.greeting  # Hi, I'm Alice
puts user.adult?    # true
```

**Accessors:**
- `getter` - read-only accessor
- `setter` - write-only accessor
- `property` - both getter and setter

**Inheritance:**

```ruby
class Admin < User
  def initialize(name : String, email : String, age : Int, @permissions : Array<String>)
    super(name, email, age)
  end

  def greeting -> String
    "[Admin] #{@name}"
  end
end
```

### Structs

Structs are lightweight, immutable value types for data without behavior:

```ruby
struct Point
  x : Int
  y : Int
end

p1 = Point{x: 10, y: 20}
p2 = Point{x: 10, y: 20}
p1 == p2  # true (value equality)
```

Structs automatically provide field accessors, value equality, and can be used as map keys.

### Interfaces

Interfaces define capabilities. Types satisfy interfaces implicitly (structural typing).

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

announce(Dog.new)  # Woof!
announce(Cat.new)  # Meow!
```

Use `implements` for compile-time verification (optional but documents intent):

```ruby
class Robot implements Speaker
  def speak -> String
    "Beep boop"
  end
end
```

### Modules

Modules share behavior across classes:

```ruby
module Greetable
  def greet -> String
    "Hello, I'm #{@name}"
  end
end

class User
  include Greetable
  property name : String

  def initialize(@name : String)
  end
end

puts User.new("Alice").greet  # Hello, I'm Alice
```

Modules also serve as namespaces:

```ruby
module Http
  class Client
    def get(url : String) -> Response
      # ...
    end
  end
end

client = Http::Client.new
```

### Visibility

By default, everything is internal to the package. Use `pub` to export:

```ruby
# Internal - only usable within this package
def helper(x : Int) -> Int
  x * 2
end

# Exported - callable from other packages
pub def double(x : Int) -> Int
  helper(x)
end

pub class Counter
  pub def inc
    @count += 1
  end

  def reset  # internal method
    @count = 0
  end
end
```

### Error Handling

Rugby uses Go's error-as-value pattern with concise syntax. The error type is `Error` (PascalCase).

**The `!` operator** propagates errors to the caller:

```ruby
def load_config(path : String) -> (Config, Error)
  content = file.read(path)!          # returns error if file.read fails
  config = json.parse(content)!       # returns error if json.parse fails
  config, nil
end
```

**The `rescue` keyword** provides fallbacks:

```ruby
# Inline default
port = env.get("PORT").to_i rescue 8080

# Block form with logging
config = json.parse(text) rescue do
  puts "Invalid JSON, using defaults"
  default_config
end

# With error binding
data = file.read(path) rescue => err do
  log.warn("read failed: #{err}")
  default_data
end
```

**Explicit handling** when you need full control:

```ruby
result, err = some_operation()
if err != nil
  log.error("Operation failed: #{err}")
  return nil, err
end
```

### Optionals

Optional types use `T?` and require explicit handling.

```ruby
def find_user(id : Int) -> User?
  return nil if id < 0
  users[id]
end
```

**Nil coalescing** with `??`:

```ruby
name = user&.name ?? "Anonymous"
port = config.get("port") ?? 8080
```

**Safe navigation** with `&.`:

```ruby
city = user&.address&.city ?? "Unknown"
```

**Pattern matching** with `if let`:

```ruby
if let user = find_user(42)
  puts user.name
else
  puts "User not found"
end
```

**Tuple unpacking:**

```ruby
user, ok = find_user(42)
if ok
  puts user.name
end
```

### Strings

**Interpolation** (double quotes):

```ruby
message = "Hello, #{user.name}! You have #{count} messages."
```

**Literal strings** (single quotes - no interpolation):

```ruby
pattern = '#{not interpolated}'
```

**Heredocs** for multi-line strings:

```ruby
query = <<~SQL
  SELECT *
  FROM users
  WHERE active = true
SQL
```

### Ranges

Ranges represent sequences of integers:

```ruby
1..5     # inclusive: 1, 2, 3, 4, 5
1...5    # exclusive: 1, 2, 3, 4

for i in 0..10
  puts i
end

(1..100).include?(50)  # true
nums = (1..5).to_a     # [1, 2, 3, 4, 5]
```

### Go Interop

Import and call Go packages directly. Rugby's snake_case automatically maps to Go's CamelCase:

```ruby
import "fmt"
import "strings"
import "os"

fmt.Println("Hello from Go")

# Rugby style (snake_case) - preferred
upper = strings.to_upper("hello")      # calls strings.ToUpper
data = os.read_file("config.txt")!     # calls os.ReadFile

# Go style also works
upper = strings.ToUpper("hello")
```

**Resource cleanup with `defer`:**

```ruby
import "os"
import "io"

def process_file(path : String) -> Error
  f = os.open(path)!
  defer f.close()  # runs when function exits

  data = io.read_all(f)!
  process(data)
end
```

### Concurrency

Rugby provides Ruby-like ergonomics for Go's concurrency.

**Spawn and await** for value-returning tasks:

```ruby
t1 = spawn { fetch_user(1) }
t2 = spawn { fetch_user(2) }

user1 = await t1
user2 = await t2
```

**Structured concurrency** with `concurrently`:

```ruby
concurrently -> (scope) do
  a = scope.spawn { fetch_a() }
  b = scope.spawn { fetch_b() }

  result_a = await(a)!
  result_b = await(b)!

  process(result_a, result_b)
end
# All tasks guaranteed complete or cancelled here
```

**Channels** for communication:

```ruby
ch = Chan<Int>.new(10)  # buffered channel
ch << 42                # send
value = ch.receive      # blocking receive

for msg in ch           # iterate until closed
  process(msg)
end
```

**Goroutines** for fire-and-forget:

```ruby
go do
  background_work()
end
```

## Standard Library

Rugby includes a standard library with Ruby-like APIs.

| Package | Description |
|---------|-------------|
| `rugby/http` | HTTP client |
| `rugby/http_server` | HTTP server |
| `rugby/json` | JSON parsing/generation |
| `rugby/file` | File I/O |
| `rugby/env` | Environment variables |
| `rugby/time` | Time and duration |
| `rugby/path` | Path manipulation |
| `rugby/regex` | Regular expressions |
| `rugby/shell` | Command execution |
| `rugby/csv` | CSV reading/writing |
| `rugby/log` | Structured logging |
| `rugby/uuid` | UUID generation |
| `rugby/crypto` | Cryptographic utilities |
| `rugby/base64` | Base64 encoding |
| `rugby/url` | URL parsing |

### Examples

**HTTP request:**

```ruby
import "rugby/http"

resp = http.get("https://api.example.com/users")!
if resp.ok?
  users = resp.json_array!
  users.each -> (user) { puts user["name"] }
end
```

**File operations:**

```ruby
import "rugby/file"
import "rugby/json"

content = file.read("config.json")!
config = json.parse(content)!
puts config["setting"]
```

**Environment and configuration:**

```ruby
import "rugby/env"

port = env.fetch("PORT", "8080")
debug = env.fetch("DEBUG", "false") == "true"
```

**Running commands:**

```ruby
import "rugby/shell"

output = shell.run("git status --short")!
puts output

if shell.exists?("docker")
  shell.run("docker build -t myapp .")!
end
```

**Time operations:**

```ruby
import "rugby/time"

start = time.now
do_work()
elapsed = time.since(start)
puts "Took #{elapsed.string}"

tomorrow = time.now.add(time.hours(24))
```

## Development

```bash
# Build
make build

# Run tests
make test

# Run linters
make lint

# Format code
make fmt

# Run all checks
make check
```

## More Resources

- [Language Specification](spec.md) - Complete language reference
- [Idiomatic Rugby](docs/idioms.md) - Style guide and best practices
- [Standard Library](stdlib/) - Package documentation
- [Examples](examples/) - Sample programs
