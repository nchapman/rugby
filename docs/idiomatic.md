# Idiomatic Rugby

Rugby brings Ruby's expressiveness to Go's performance. This guide establishes what natural, idiomatic Rugby code looks like.

## Naming

Rugby uses Ruby-style naming throughout:

```ruby
# Functions and methods: snake_case
def fetch_user(id : Int) -> User?
  db.find_by_id(id)
end

# Predicates end with ?
def valid?
  @name.length > 0
end

# Types and classes: CamelCase
class UserAccount
  property email : String
end

# Variables: snake_case
current_user = find_user(id)
total_count = items.length
```

When calling Go interop, Rugby automatically maps `snake_case` to Go's `CamelCase`:

```ruby
import net/http

# These are equivalent - use snake_case
resp = http.get(url)      # calls http.Get
body = io.read_all(resp)  # calls io.ReadAll
```

## Embrace Lambdas

Lambdas are Rugby's bread and butter for transformation. Prefer them over manual loops:

```ruby
# Good - expressive and clear
names = users.map -> (u) { u.name }
active = users.select -> (u) { u.active? }
total = orders.reduce(0) -> (sum, o) { sum + o.amount }

# Avoid - verbose and imperative
names = []
for u in users
  names << u.name
end
```

Use `do...end` for multi-line lambdas, braces for single-line:

```ruby
# Single line - braces
squares = nums.map -> (n) { n * n }

# Multi-line - do...end
results = items.map -> (item) do
  processed = transform(item)
  validate(processed)
end
```

The symbol-to-proc shorthand keeps things concise:

```ruby
names = users.map(&:name)           # same as: users.map -> (u) { u.name }
active = users.select(&:active?)    # same as: users.select -> (u) { u.active? }
emails = users.reject(&:empty?)
```

### Lambdas vs For Loops

Lambdas are for data transformation. `return` inside a lambda returns from the lambda, not the enclosing function. Use `for` loops when you need control flow:

```ruby
# Lambdas for transformation
active = users.select -> (u) { u.active? }
admin = users.find -> (u) { u.admin? }

# For loops when you need break/next/return
def find_first_problem(items : Array<Item>) -> Item?
  for item in items
    next if item.ok?
    return item  # returns from function
  end
  nil
end
```

## Optionals Over Nil Checks

Rugby's optional type (`T?`) has operators that eliminate nil-checking boilerplate:

```ruby
# Good - nil coalescing
name = user&.name ?? "Anonymous"
port = config.get("port") ?? 8080

# Good - safe navigation chains
city = user&.address&.city ?? "Unknown"

# Good - if let for conditional binding
if let user = find_user(id)
  puts "Found: #{user.name}"
end

# Avoid - manual nil checks
user, ok = find_user(id)
if ok
  name = user.name
else
  name = "Anonymous"
end
```

## Error Handling

Rugby offers three patterns. Choose based on context:

### Propagate with `!`

When errors should bubble up to the caller:

```ruby
def load_config(path : String) -> (Config, error)
  data = file.read(path)!
  json.parse(data)!
end
```

### Recover with `rescue`

When you have a sensible default or recovery strategy:

```ruby
# Inline default
timeout = config.get("timeout").to_i rescue 30

# Block form for complex recovery
data = file.read(path) rescue do
  log.warn "config missing, using defaults"
  default_config
end

# With error binding
result = fetch_data(url) rescue => err do
  log.error "fetch failed: #{err}"
  cached_data
end
```

### Explicit handling

When you need fine-grained control:

```ruby
data, err = file.read(path)
if err != nil
  if errors.is?(err, os.ErrNotExist)
    return create_default
  end
  return nil, err
end
```

## Calling Convention

Parentheses are optional. Use them with arguments, drop them for no-arg calls:

```ruby
# With arguments - use parens
process(item)
users.each -> (u) { send_email(u) }
notify_subscribers(event)

# No arguments - drop parens
result = expensive_calculation
data = resp.json
config = load_defaults

# Exception: output functions stay paren-free
puts "hello"
print "loading..."
p debug_value
log.info "server started"
```

This keeps code unambiguous while staying clean.

Map literals need parens to distinguish from lambdas:

```ruby
send_request({method: "POST", body: data})  # map argument
send_request -> { do_something() }          # lambda argument
```

## Statement Modifiers

Put simple conditions at the end for readability:

```ruby
return nil if id < 0
puts "invalid" unless valid?
break if done
next unless item.active?
```

Reserve block form for complex conditions or multiple statements:

```ruby
if user.admin? && user.verified? && !maintenance_mode?
  grant_access(user)
  log_admin_action(user, action)
end
```

## Concurrency

Rugby makes concurrency approachable. Choose the right tool:

### Fire and forget

```ruby
go notify_subscribers(event)

go do
  sync_to_backup(data)
end
```

### Get a result back

```ruby
task = spawn { expensive_calculation }
# ... do other work ...
result = await task
```

### Parallel operations with cleanup

```ruby
concurrently -> (scope) do
  user_task = scope.spawn { fetch_user(id) }
  orders_task = scope.spawn { fetch_orders(id) }

  user = await(user_task)
  orders = await(orders_task)

  build_profile(user, orders)
end
# All tasks guaranteed complete or cancelled here
```

### Channels for communication

```ruby
ch = Chan<Event>.new(100)

# Producer
go do
  for e in events
    ch << e
  end
  ch.close
end

# Consumer
for event in ch
  process(event)
end
```

## String Interpolation

Always prefer interpolation over concatenation:

```ruby
# Good
message = "Hello, #{user.name}! You have #{count} messages."
path = "#{base_dir}/#{filename}"

# Avoid
message = "Hello, " + user.name + "! You have " + count.to_s + " messages."
```

Heredocs for multi-line strings:

```ruby
query = <<~SQL
  SELECT *
  FROM users
  WHERE active = true
  ORDER BY created_at DESC
SQL
```

## Classes

### Defining Fields

Fields can be declared three ways:

```ruby
class User
  @role : String                    # explicit declaration

  def initialize(@name : String)    # parameter promotion (declares + assigns)
    @created_at = time.now          # inferred from initialize
  end
end
```

Parameter promotion (`@name : String` in the signature) is the most concise - it declares the field and assigns the argument in one go.

### Accessors

Use `getter`, `setter`, and `property` to generate accessor methods:

```ruby
class User
  getter id : Int           # def id -> Int
  setter status : String    # def status=(v : String)
  property name : String    # both getter and setter
end

user = User.new(1, "active", "Alice")
puts user.name              # getter
user.name = "Bob"           # setter
```

### Visibility

Use `pub` to export classes and methods:

```ruby
pub class User              # exported (Go: User)
  pub getter name : String  # exported accessor

  def internal_method       # package-private (Go: internalMethod)
    # ...
  end
end
```

### Inheritance and Interfaces

```ruby
class Admin < User                    # single inheritance
  def initialize(name : String, @permissions : Array<String>)
    super(name)                       # call parent initializer
  end

  def display_name                    # override parent method
    "[Admin] #{@name}"
  end
end

class Worker implements Runnable      # explicit interface (optional)
  def run
    # ...
  end
end
```

Rugby uses structural typing - classes satisfy interfaces automatically if they have matching methods. The `implements` declaration is optional but enables compile-time checking.

### Self

Inside methods, `self` refers to the receiver. It's optional when calling other methods:

```ruby
class Counter
  def initialize
    @count = 0
  end

  def inc
    @count += 1
    self              # return self for chaining
  end

  def reset
    @count = 0
    inc               # same as self.inc
  end
end
```

## Putting It Together

Here's a complete example showing idiomatic Rugby:

```ruby
import rugby/http
import rugby/json
import rugby/file

class ApiClient
  def initialize(@base_url : String, @timeout : Int = 30)
  end

  def fetch_users -> (Array<User>, error)
    resp = http.get("#{@base_url}/users")!
    data = resp.json!

    data.map -> (raw) do
      User.new(
        name: raw["name"].as(String),
        email: raw["email"].as(String)
      )
    end
  end
end

def main
  client = ApiClient.new("https://api.example.com")

  users = client.fetch_users rescue do
    puts "API unavailable, loading from cache"
    load_cached_users
  end

  active = users.select(&:active?)

  puts "Found #{active.length} active users:"
  active.each -> (u) { puts "  - #{u.name}" }  # can't use &: here (needs arg)
end
```

## Summary

Idiomatic Rugby:

- Uses `snake_case` for functions, methods, and variables
- Embraces lambdas and `&:method` for transformation
- Uses `for` loops when you need `break`, `next`, or `return`
- Leverages `??` and `&.` for optional handling
- Uses `!` and `rescue` for clean error handling
- Prefers string interpolation
- Takes advantage of statement modifiers
- Keeps concurrency simple with `spawn`/`await` and `concurrently`
