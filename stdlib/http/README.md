# rugby/http

A simple HTTP client for Rugby programs.

## Import

```ruby
import rugby/http
```

## Quick Start

```ruby
import rugby/http

# GET request
resp = http.Get("https://api.example.com/users")!
puts resp.Body()

# POST with JSON
resp = http.Post("https://api.example.com/users", http.Options{
  JSON: {name: "Alice", age: 30}
})!

# Check response
if resp.Ok()
  data = resp.JSON()!
  puts data["id"]
end
```

## Functions

### http.Get(url, options?) -> (Response, error)

Performs an HTTP GET request.

```ruby
resp = http.Get("https://example.com")!
resp = http.Get(url, http.Options{Headers: {"Authorization": "Bearer token"}})!
```

### http.Post(url, options?) -> (Response, error)

Performs an HTTP POST request.

```ruby
# POST with JSON body
resp = http.Post(url, http.Options{JSON: {name: "Alice"}})!

# POST with form data
resp = http.Post(url, http.Options{Form: {"username": "alice", "password": "secret"}})!

# POST with raw body
resp = http.Post(url, http.Options{Body: some_bytes})!
```

### http.Put(url, options?) -> (Response, error)

Performs an HTTP PUT request.

```ruby
resp = http.Put(url, http.Options{JSON: {name: "Updated"}})!
```

### http.Patch(url, options?) -> (Response, error)

Performs an HTTP PATCH request.

```ruby
resp = http.Patch(url, http.Options{JSON: {status: "active"}})!
```

### http.Delete(url, options?) -> (Response, error)

Performs an HTTP DELETE request.

```ruby
resp = http.Delete("https://api.example.com/users/123")!
```

### http.Head(url, options?) -> (Response, error)

Performs an HTTP HEAD request (headers only, no body).

```ruby
resp = http.Head(url)!
puts resp.Headers["Content-Length"]
```

## Options

The `http.Options` struct configures requests:

| Field | Type | Description |
|-------|------|-------------|
| `Headers` | `Map[String, String]` | Custom request headers |
| `JSON` | `any` | JSON body (auto-sets Content-Type) |
| `Form` | `Map[String, String]` | Form data (application/x-www-form-urlencoded) |
| `Body` | `Bytes` | Raw body bytes |
| `Timeout` | `Duration` | Request timeout (default: 30s) |

```ruby
opts = http.Options{
  Headers: {"Authorization": "Bearer #{token}", "X-Custom": "value"},
  Timeout: 10 * time.Second
}
resp = http.Get(url, opts)!
```

## Response

The `Response` struct provides access to the HTTP response:

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `Status` | `Int` | HTTP status code (200, 404, etc.) |
| `StatusText` | `String` | Full status line ("200 OK") |
| `Headers` | `Map[String, String]` | Response headers |

### Methods

#### resp.Ok() -> Bool

Returns `true` if status code is 2xx (success).

```ruby
if resp.Ok()
  puts "Success!"
end
```

#### resp.Body() -> String

Returns the response body as a string.

```ruby
html = resp.Body()
```

#### resp.Bytes() -> Bytes

Returns the response body as raw bytes.

```ruby
data = resp.Bytes()
```

#### resp.JSON() -> (Map[String, any], error)

Parses the response body as JSON into a map.

```ruby
data = resp.JSON()!
puts data["name"]
```

#### resp.JSONArray() -> (Array[any], error)

Parses the response body as a JSON array.

```ruby
items = resp.JSONArray()!
for item in items
  puts item["id"]
end
```

#### resp.JSONInto(target) -> error

Parses the response body into a struct.

```ruby
user = User{}
resp.JSONInto(user)!
```

## Error Handling

All request functions return `(Response, error)`. Use `!` to propagate errors or `rescue` for fallbacks:

```ruby
# Propagate error
resp = http.Get(url)!

# Provide fallback
resp = http.Get(url) rescue do
  puts "Request failed, using cached data"
  cached_response
end

# Handle explicitly
resp, err = http.Get(url)
if err != nil
  puts "Error: #{err}"
  return
end
```

## Examples

### Fetch JSON API

```ruby
import rugby/http

resp = http.Get("https://jsonplaceholder.typicode.com/users/1")!
if resp.Ok()
  user = resp.JSON()!
  puts "Name: #{user["name"]}"
  puts "Email: #{user["email"]}"
end
```

### POST JSON Data

```ruby
import rugby/http

resp = http.Post("https://api.example.com/users", http.Options{
  JSON: {name: "Alice", email: "alice@example.com"},
  Headers: {"Authorization": "Bearer #{api_token}"}
})!

if resp.Ok()
  result = resp.JSON()!
  puts "Created user #{result["id"]}"
else
  puts "Error: #{resp.Status}"
end
```

### Download File

```ruby
import rugby/http
import rugby/file

resp = http.Get("https://example.com/data.csv")!
file.Write("data.csv", resp.Body())!
```
