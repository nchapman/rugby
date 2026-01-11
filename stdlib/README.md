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

## TODO: Planned Packages

### Tier 1: Essential

- [ ] `rugby/file` - File I/O operations
  - `file.Read(path)`, `file.ReadBytes(path)`, `file.ReadLines(path)`
  - `file.Write(path, content)`, `file.Append(path, content)`
  - `file.Exists?(path)`, `file.Size(path)`, `file.Directory?(path)`

- [ ] `rugby/env` - Environment variable access
  - `env.Get(key)` returns `String?`
  - `env.Fetch(key, default)` returns `String`
  - `env.Set(key, value)`, `env.Delete(key)`
  - `env.All()` returns `Map[String, String]`

- [ ] `rugby/path` - Path manipulation (pure functions, no I/O)
  - `path.Join(parts...)`, `path.Split(path)`
  - `path.Dirname(path)`, `path.Basename(path)`, `path.Extname(path)`
  - `path.Absolute?(path)`, `path.Relative(base, target)`

### Tier 2: Common

- [ ] `rugby/time` - Time and duration
  - `time.Now()`, `time.Parse(str, format)`, `time.Unix(seconds)`
  - `t.Format(layout)`, `t.Add(duration)`, `t.Sub(other)`
  - `time.Since(t)`, `time.Until(t)`
  - Duration helpers: `time.Seconds(n)`, `time.Minutes(n)`, etc.

- [ ] `rugby/regex` - Regular expressions
  - `regex.Match?(pattern, str)`, `regex.Find(pattern, str)`
  - `regex.FindAll(pattern, str)`, `regex.Replace(str, pattern, replacement)`
  - `regex.Split(str, pattern)`

- [ ] `rugby/base64` - Base64 encoding/decoding
  - `base64.Encode(bytes)`, `base64.Decode(str)`
  - `base64.URLEncode(bytes)`, `base64.URLDecode(str)`

### Tier 3: Nice to Have

- [ ] `rugby/shell` - Command execution
  - `shell.Run(cmd)` returns `(String, error)`
  - `shell.Exec(cmd, args)` returns `Result{Stdout, Stderr, ExitCode}`
  - `shell.Which(name)` returns `String?`

- [ ] `rugby/crypto` - Cryptographic utilities
  - `crypto.SHA256(data)`, `crypto.MD5(data)`
  - `crypto.RandomBytes(n)`

- [ ] `rugby/url` - URL parsing and building
  - `url.Parse(str)`, `url.Join(base, path)`
  - `url.QueryEncode(params)`, `url.QueryDecode(str)`

- [ ] `rugby/template` - Simple string templating
  - `template.Render(tmpl, data)`
