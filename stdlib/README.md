# Rugby Standard Library

This directory contains Rugby's standard library packages. These provide Ruby-like ergonomics for common tasks while compiling to idiomatic Go.

## Available Packages

| Package | Description | Import |
|---------|-------------|--------|
| [http](./http/) | HTTP client for web requests | `import rugby/http` |
| [json](./json/) | JSON parsing and generation | `import rugby/json` |
| [file](./file/) | File I/O operations | `import rugby/file` |
| [env](./env/) | Environment variable access | `import rugby/env` |
| [path](./path/) | Path manipulation | `import rugby/path` |
| [time](./time/) | Time and duration | `import rugby/time` |
| [regex](./regex/) | Regular expressions | `import rugby/regex` |
| [base64](./base64/) | Base64 encoding/decoding | `import rugby/base64` |
| [shell](./shell/) | Command execution | `import rugby/shell` |
| [crypto](./crypto/) | Cryptographic utilities | `import rugby/crypto` |
| [url](./url/) | URL parsing and building | `import rugby/url` |

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

## TODO: Planned Packages

### Tier 2: High Value

- [ ] `rugby/uuid` - UUID generation and parsing
  - `uuid.V4()` returns random UUID string
  - `uuid.Parse(str)` validates and normalizes
  - `uuid.Valid?(str)` checks format

- [ ] `rugby/csv` - CSV reading and writing
  - `csv.Parse(str)` returns `Array[Array[String]]`
  - `csv.ParseWithHeaders(str)` returns `Array[Map[String, String]]`
  - `csv.Generate(rows)` returns CSV string

- [ ] `rugby/log` - Structured logging with levels
  - `log.Info(msg, fields)`, `log.Error(msg, fields)`
  - `log.SetLevel(level)` controls output
  - JSON and text output formats

### Tier 3: Nice to Have

- [ ] `rugby/template` - Simple string templating
  - `template.Render(tmpl, data)`
