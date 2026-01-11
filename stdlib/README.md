# Rugby Standard Library

This directory contains Rugby's standard library packages. These provide Ruby-like ergonomics for common tasks while compiling to idiomatic Go.

## Available Packages

| Package | Description | Import |
|---------|-------------|--------|
| [http](./http/) | HTTP client for web requests | `import rugby/http` |
| [json](./json/) | JSON parsing and generation | `import rugby/json` |

## Usage

Unlike kernel functions (`puts`, `gets`, etc.) which are available automatically, stdlib packages require explicit imports:

```ruby
import rugby/http
import rugby/json

resp = http.Get("https://api.example.com/data")!
data = resp.JSON()!
puts data["message"]
```

## Design Principles

1. **Explicit imports** - No magic; dependencies are visible at the top of the file
2. **Error returns** - Functions return `(T, error)` following Go conventions
3. **Ruby-like API** - Method names and patterns familiar to Ruby developers
4. **Thin wrappers** - Minimal abstraction over Go's stdlib

## Coming Soon

- `rugby/file` - File I/O operations
- `rugby/env` - Environment variable access
- `rugby/path` - Path manipulation utilities
- `rugby/time` - Time and duration handling
