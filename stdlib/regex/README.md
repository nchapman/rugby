# rugby/regex

Regular expression operations for Rugby programs.

## Import

```ruby
import rugby/regex
```

## Quick Start

```ruby
import rugby/regex

# Check for a match
if regex.Match(`\d+`, "abc123")
  puts "Found digits!"
end

# Find first match
if let match = regex.Find(`\w+`, "hello world")
  puts match  # hello
end

# Find all matches
matches = regex.FindAll(`\d+`, "a1b2c3")
# ["1", "2", "3"]

# Replace matches
result = regex.Replace("hello world", `\w+`, "X")
# "X X"
```

## Functions

### Matching

#### regex.Match(pattern, str) -> Bool

Reports whether the string contains any match of the pattern.

```ruby
regex.Match(`\d+`, "abc123")     # true
regex.Match(`\d+`, "abc")        # false
regex.Match(`^hello`, "hello!")  # true
```

#### regex.MustMatch(pattern, str) -> Bool

Like `Match`, but panics if the pattern is invalid.

```ruby
regex.MustMatch(`\d+`, "abc123")  # true
regex.MustMatch(`[invalid`, "x") # panics!
```

### Finding

#### regex.Find(pattern, str) -> (String, Bool)

Returns the first match and whether one was found.

```ruby
match, ok = regex.Find(`\d+`, "abc123def")
# match = "123", ok = true

match, ok = regex.Find(`\d+`, "abc")
# match = "", ok = false

# Using if let
if let match = regex.Find(`\w+@\w+`, text)
  puts "Found email-like: #{match}"
end
```

#### regex.FindAll(pattern, str) -> Array[String]

Returns all matches.

```ruby
matches = regex.FindAll(`\d+`, "a1b2c3")
# ["1", "2", "3"]

words = regex.FindAll(`\w+`, "hello, world!")
# ["hello", "world"]
```

#### regex.FindGroups(pattern, str) -> Array[String]

Returns the first match and its capture groups.

```ruby
groups = regex.FindGroups(`(\w+)@(\w+)\.(\w+)`, "test@example.com")
# ["test@example.com", "test", "example", "com"]
# groups[0] = full match
# groups[1] = first capture group
# groups[2] = second capture group
```

#### regex.FindAllGroups(pattern, str) -> Array[Array[String]]

Returns all matches with their capture groups.

```ruby
groups = regex.FindAllGroups(`(\w+)=(\d+)`, "a=1 b=2 c=3")
# [["a=1", "a", "1"], ["b=2", "b", "2"], ["c=3", "c", "3"]]
```

### Replacing

#### regex.Replace(str, pattern, replacement) -> String

Replaces all matches with the replacement string. Supports backreferences like `$1`, `$2`.

```ruby
regex.Replace("hello world", `\w+`, "X")
# "X X"

regex.Replace("hello world", `(\w+) (\w+)`, "$2 $1")
# "world hello"
```

#### regex.ReplaceFirst(str, pattern, replacement) -> String

Replaces only the first match. Supports backreferences.

```ruby
regex.ReplaceFirst("a1b2c3", `\d+`, "X")
# "aXb2c3"

regex.ReplaceFirst("hello world", `(\w+)`, "[$1]")
# "[hello] world"
```

### Splitting

#### regex.Split(str, pattern) -> Array[String]

Splits the string by the pattern.

```ruby
regex.Split("a1b2c3", `\d`)
# ["a", "b", "c", ""]

regex.Split("one,two;three", `[,;]`)
# ["one", "two", "three"]
```

### Escaping

#### regex.Escape(str) -> String

Escapes all regex metacharacters in the string.

```ruby
regex.Escape("hello.world")   # "hello\.world"
regex.Escape("a+b*c?")        # "a\+b\*c\?"
regex.Escape("[test]")        # "\[test\]"
```

### Compiled Patterns

For repeated use, compile the pattern once:

#### regex.Compile(pattern) -> (Regex, error)

Compiles a pattern into a reusable Regex object.

```ruby
re, err = regex.Compile(`\d+`)
if err != nil
  puts "Invalid pattern: #{err}"
  return
end

re.Match("abc123")  # true
```

#### regex.MustCompile(pattern) -> Regex

Like `Compile`, but panics on invalid patterns.

```ruby
re = regex.MustCompile(`\w+`)
matches = re.FindAll("hello world")
```

### Regex Methods

The compiled `Regex` type has methods matching the standalone functions:

```ruby
re = regex.MustCompile(`(\w+)`)

re.Match("hello")              # true
re.Find("hello world")         # ("hello", true)
re.FindAll("hello world")      # ["hello", "world"]
re.FindGroups("hello")         # ["hello", "hello"]
re.Replace("a b c", "X")       # "X X X"
re.ReplaceFirst("a b c", "X")  # "X b c"
re.Split("a1b2c3")             # (depends on pattern)
re.Pattern()                   # returns the pattern string
```

## Pattern Syntax

Rugby uses Go's RE2 regular expression syntax:

| Pattern | Matches |
|---------|---------|
| `.` | Any character (except newline) |
| `*` | Zero or more of preceding |
| `+` | One or more of preceding |
| `?` | Zero or one of preceding |
| `^` | Start of string |
| `$` | End of string |
| `[abc]` | Any character in set |
| `[^abc]` | Any character not in set |
| `\d` | Digit (0-9) |
| `\w` | Word character (a-z, A-Z, 0-9, _) |
| `\s` | Whitespace |
| `(...)` | Capture group |
| `(?:...)` | Non-capturing group |
| `a\|b` | Alternation (a or b) |

## Error Handling

Functions with invalid patterns return safe defaults instead of panicking:

```ruby
# Invalid pattern returns false
regex.Match(`[invalid`, "test")  # false

# Invalid pattern returns no match
match, ok = regex.Find(`[invalid`, "test")  # "", false

# Invalid pattern returns original string
regex.Replace("test", `[invalid`, "X")  # "test"
```

Use `MustMatch` or `MustCompile` when you want to panic on invalid patterns.

## Examples

### Validate Email

```ruby
import rugby/regex

def valid_email?(email : String) -> Bool
  regex.Match(`^[\w.+-]+@[\w.-]+\.\w{2,}$`, email)
end

valid_email?("user@example.com")  # true
valid_email?("invalid")           # false
```

### Extract Data

```ruby
import rugby/regex

text = "Order #12345 placed on 2024-01-15"

# Extract order number
if let order = regex.Find(`#(\d+)`, text)
  puts "Order: #{order}"
end

# Extract all numbers
numbers = regex.FindAll(`\d+`, text)
# ["12345", "2024", "01", "15"]
```

### Parse Log Line

```ruby
import rugby/regex

def parse_log(line : String) -> Map[String, String]
  groups = regex.FindGroups(
    `^\[(\d{4}-\d{2}-\d{2})\] (\w+): (.*)$`,
    line
  )

  if groups.length < 4
    return {}
  end

  {
    "date": groups[1],
    "level": groups[2],
    "message": groups[3]
  }
end

log = parse_log("[2024-01-15] ERROR: Connection failed")
puts log["level"]  # ERROR
```

### Sanitize Input

```ruby
import rugby/regex

def sanitize(input : String) -> String
  # Remove HTML tags
  no_tags = regex.Replace(input, `<[^>]*>`, "")
  # Collapse whitespace
  regex.Replace(no_tags, `\s+`, " ").strip
end

sanitize("<b>Hello</b>  World")  # "Hello World"
```

### Tokenizer

```ruby
import rugby/regex

def tokenize(code : String) -> Array[Map[String, String]]
  patterns = [
    {type: "NUMBER", pattern: `\d+`},
    {type: "IDENT", pattern: `[a-zA-Z_]\w*`},
    {type: "OP", pattern: `[+\-*/=]`}
  ]

  tokens = []
  for p in patterns
    matches = regex.FindAll(p["pattern"], code)
    for m in matches
      tokens.push({type: p["type"], value: m})
    end
  end
  tokens
end
```
