# rugby/json

JSON parsing and generation for Rugby programs.

## Import

```ruby
import rugby/json
```

## Quick Start

```ruby
import rugby/json

# Parse JSON string
data = json.Parse('{"name": "Alice", "age": 30}')!
puts data["name"]  # Alice

# Generate JSON string
user = {name: "Bob", score: 100}
str = json.Generate(user)!
puts str  # {"name":"Bob","score":100}

# Pretty print
pretty = json.Pretty(user)!
puts pretty
# {
#   "name": "Bob",
#   "score": 100
# }
```

## Functions

### Parsing

#### json.Parse(str) -> (Map[String, any], error)

Parses a JSON string into a map.

```ruby
data = json.Parse('{"name": "Alice", "active": true}')!
puts data["name"]    # Alice
puts data["active"]  # true
```

#### json.ParseArray(str) -> (Array[any], error)

Parses a JSON string into an array.

```ruby
items = json.ParseArray('[1, 2, 3, "four"]')!
puts items[0]  # 1
puts items[3]  # four
```

#### json.ParseBytes(bytes) -> (Map[String, any], error)

Parses JSON bytes into a map.

```ruby
data = json.ParseBytes(response_bytes)!
```

#### json.ParseInto(str, target) -> error

Parses a JSON string into a struct.

```ruby
class User
  property name : String
  property age : Int
end

user = User{}
json.ParseInto('{"name": "Alice", "age": 30}', user)!
puts user.name  # Alice
```

### Generation

#### json.Generate(value) -> (String, error)

Converts a value to a compact JSON string.

```ruby
data = {name: "Alice", scores: [95, 87, 92]}
str = json.Generate(data)!
# {"name":"Alice","scores":[95,87,92]}
```

#### json.GenerateBytes(value) -> (Bytes, error)

Converts a value to JSON bytes.

```ruby
bytes = json.GenerateBytes(data)!
```

#### json.Pretty(value) -> (String, error)

Converts a value to a pretty-printed JSON string with indentation.

```ruby
data = {name: "Alice", address: {city: "NYC", zip: "10001"}}
str = json.Pretty(data)!
# {
#   "name": "Alice",
#   "address": {
#     "city": "NYC",
#     "zip": "10001"
#   }
# }
```

#### json.PrettyBytes(value) -> (Bytes, error)

Converts a value to pretty-printed JSON bytes.

```ruby
bytes = json.PrettyBytes(data)!
```

### Validation

#### json.Valid(str) -> Bool

Returns `true` if the string is valid JSON.

```ruby
if json.Valid(input)
  data = json.Parse(input)!
else
  puts "Invalid JSON"
end
```

#### json.ValidBytes(bytes) -> Bool

Returns `true` if the bytes are valid JSON.

```ruby
if json.ValidBytes(raw_data)
  # safe to parse
end
```

## Type Mapping

| JSON Type | Rugby Type |
|-----------|------------|
| `object` | `Map[String, any]` |
| `array` | `Array[any]` |
| `string` | `String` |
| `number` | `Float` (always float64) |
| `boolean` | `Bool` |
| `null` | `nil` |

Note: JSON numbers are always parsed as `Float`. Cast to `Int` if needed:

```ruby
data = json.Parse('{"count": 42}')!
count = data["count"].to_i  # Convert to Int
```

## Error Handling

All parsing and generation functions return errors. Use `!` to propagate or `rescue` for fallbacks:

```ruby
# Propagate error
data = json.Parse(str)!

# Provide default on error
data = json.Parse(str) rescue {}

# Handle explicitly
data, err = json.Parse(str)
if err != nil
  puts "Invalid JSON: #{err}"
  return
end
```

## Examples

### Parse API Response

```ruby
import rugby/http
import rugby/json

resp = http.Get("https://api.example.com/user")!
user = resp.JSON()!

puts "Welcome, #{user["name"]}!"
```

### Build and Send JSON

```ruby
import rugby/http
import rugby/json

payload = {
  action: "create",
  data: {
    name: "New Item",
    tags: ["important", "urgent"]
  }
}

# Using http.Options{JSON: ...} is preferred, but you can also:
body = json.Generate(payload)!
resp = http.Post(url, http.Options{
  Body: body.bytes,
  Headers: {"Content-Type": "application/json"}
})!
```

### Config File

```ruby
import rugby/file
import rugby/json

# Read and parse config
content = file.Read("config.json")!
config = json.Parse(content)!

port = config["port"] ?? 8080
debug = config["debug"] ?? false

puts "Starting on port #{port}"
```

### Pretty Print for Debugging

```ruby
import rugby/json

data = fetch_complex_data()!
puts json.Pretty(data)!  # Readable output for debugging
```
