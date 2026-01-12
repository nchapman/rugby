# rugby/path

Path manipulation for Rugby programs. These are pure functions that do not perform I/O.

## Import

```ruby
import rugby/path
```

## Quick Start

```ruby
import rugby/path

# Join path components
full = path.Join("/home", "user", "documents", "file.txt")
# /home/user/documents/file.txt

# Get directory and filename
dir = path.Dirname(full)   # /home/user/documents
name = path.Basename(full) # file.txt
ext = path.Extname(full)   # .txt

# Get filename without extension
stem = path.Stem(full)     # file
```

## Functions

### Path Construction

#### path.Join(parts...) -> String

Joins path elements into a single path.

```ruby
path.Join("a", "b", "c")        # a/b/c
path.Join("/home", "user")      # /home/user
path.Join("dir", "", "file")    # dir/file (empty elements ignored)
```

#### path.Clean(path) -> String

Returns the shortest equivalent path.

```ruby
path.Clean("a/b/../c")    # a/c
path.Clean("a/./b/c")     # a/b/c
path.Clean("a//b//c")     # a/b/c
```

### Path Decomposition

#### path.Split(path) -> (String, String)

Splits a path into directory and file components.

```ruby
dir, file = path.Split("/a/b/c.txt")
# dir = "/a/b/"
# file = "c.txt"
```

#### path.Dirname(path) -> String

Returns the directory portion of a path.

```ruby
path.Dirname("/a/b/c.txt")  # /a/b
path.Dirname("file.txt")    # .
path.Dirname("/")           # /
```

#### path.Basename(path) -> String

Returns the last element of a path.

```ruby
path.Basename("/a/b/c.txt")  # c.txt
path.Basename("/a/b/c")      # c
path.Basename("/")           # /
```

#### path.Extname(path) -> String

Returns the file extension including the dot.

```ruby
path.Extname("file.txt")      # .txt
path.Extname("file.tar.gz")   # .gz
path.Extname("file")          # "" (empty)
path.Extname(".gitignore")    # .gitignore
```

#### path.Stem(path) -> String

Returns the filename without extension.

```ruby
path.Stem("file.txt")       # file
path.Stem("file.tar.gz")    # file.tar
path.Stem("/a/b/file.txt")  # file
```

### Path Queries

#### path.Absolute(path) -> Bool

Reports whether the path is absolute.

```ruby
path.Absolute("/home/user")  # true
path.Absolute("relative")    # false
path.Absolute("./local")     # false
```

#### path.Match(pattern, name) -> Bool

Reports whether name matches the shell pattern.

```ruby
path.Match("*.txt", "file.txt")   # true
path.Match("*.txt", "file.go")    # false
path.Match("file?", "file1")      # true
path.Match("file?", "file12")     # false
```

### Path Conversion

#### path.Abs(path) -> (String, error)

Returns the absolute path.

```ruby
abs = path.Abs(".")!
puts abs  # /current/working/directory
```

#### path.Relative(base, target) -> (String, error)

Returns a relative path from base to target.

```ruby
rel = path.Relative("/a/b", "/a/b/c/d")!  # c/d
rel = path.Relative("/a/b/c", "/a/b")!    # ..
```

## Examples

### Build Output Path

```ruby
import rugby/path

def output_path(input : String, ext : String) -> String
  dir = path.Dirname(input)
  stem = path.Stem(input)
  path.Join(dir, "#{stem}#{ext}")
end

output_path("src/main.rg", ".go")  # src/main.go
```

### Find Files by Extension

```ruby
import rugby/path

def has_extension(file : String, ext : String) -> Bool
  path.Extname(file) == ext
end

files = ["main.go", "lib.go", "README.md", "test.go"]
go_files = files.select { |f| has_extension(f, ".go") }
# ["main.go", "lib.go", "test.go"]
```

### Normalize User Path

```ruby
import rugby/path
import rugby/env

def expand_path(p : String) -> String
  if p.start_with?("~/")
    home = env.Fetch("HOME", "/")
    return path.Join(home, p[2..])
  end
  path.Clean(p)
end

expand_path("~/documents")  # /home/user/documents
expand_path("./a/../b")     # b
```

### Safe Path Join

```ruby
import rugby/path

def safe_join(base : String, user_input : String) -> (String, Bool)
  # Prevent path traversal attacks
  joined = path.Join(base, user_input)
  cleaned = path.Clean(joined)

  # Ensure result is still under base
  if cleaned.start_with?(base)
    return cleaned, true
  end
  return "", false
end

safe_join("/data", "file.txt")   # ("/data/file.txt", true)
safe_join("/data", "../etc/passwd")  # ("", false)
```
