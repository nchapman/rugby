# rugby/url

URL parsing and building for Rugby programs.

## Import

```ruby
import rugby/url
```

## Quick Start

```ruby
import rugby/url

# Parse a URL
u = url.Parse("https://example.com/path?q=search")!
puts u.Host()      # example.com
puts u.Path()      # /path
puts u.Query()["q"] # search

# Build query strings
query = url.QueryEncode({"q": "hello world", "page": "1"})
puts query  # page=1&q=hello+world

# Join URL paths
full = url.Join("https://api.example.com", "v1", "users")!
puts full  # https://api.example.com/v1/users
```

## Functions

### Parsing

#### url.Parse(str) -> (URL, error)

Parses a URL string and returns a URL object.

```ruby
u = url.Parse("https://user:pass@example.com:8080/path?q=1#section")!
puts u.Scheme()    # https
puts u.Host()      # example.com:8080
puts u.Hostname()  # example.com
puts u.Port()      # 8080
puts u.Path()      # /path
puts u.RawQuery()  # q=1
puts u.Fragment()  # section
puts u.User()      # user
```

#### url.MustParse(str) -> URL

Parses a URL string, panicking on error.

```ruby
u = url.MustParse("https://example.com")
```

### URL Object Methods

#### Accessors

```ruby
u = url.MustParse("https://example.com:8080/path?a=1#top")

u.Scheme()      # "https"
u.Host()        # "example.com:8080"
u.Hostname()    # "example.com"
u.Port()        # "8080"
u.Path()        # "/path"
u.RawQuery()    # "a=1"
u.Fragment()    # "top"
u.String()      # "https://example.com:8080/path?a=1#top"
u.IsAbsolute()  # true
```

#### Query Parameters

```ruby
u = url.MustParse("https://example.com?a=1&b=2&a=3")

# Get first value for each key
q = u.Query()          # {"a": "1", "b": "2"}

# Get all values
qv = u.QueryValues()   # {"a": ["1", "3"], "b": ["2"]}
```

#### User Info

```ruby
u = url.MustParse("https://user:pass@example.com")
u.User()           # "user"
pass, ok = u.Password()  # "pass", true
```

### Building URLs

The `With*` methods return new URL objects (immutable pattern):

```ruby
u = url.MustParse("http://example.com/old")

u2 = u.WithScheme("https")
puts u2.String()  # https://example.com/old

u3 = u.WithHost("new.com:9000")
puts u3.String()  # http://new.com:9000/old

u4 = u.WithPath("/new/path")
puts u4.String()  # http://example.com/new/path

u5 = u.WithQuery({"key": "value"})
puts u5.String()  # http://example.com/old?key=value

u6 = u.WithFragment("section")
puts u6.String()  # http://example.com/old#section
```

#### url.Resolve(ref) -> (URL, error)

Resolves a relative reference against a base URL.

```ruby
base = url.MustParse("https://example.com/a/b")
u = base.Resolve("../c")!
puts u.String()  # https://example.com/c
```

### Joining Paths

#### url.Join(base, paths...) -> (String, error)

Joins a base URL with path segments.

**Note:** Paths are cleaned (trailing slashes removed, `.` and `..` resolved).

```ruby
full = url.Join("https://api.example.com", "v1", "users")!
puts full  # https://api.example.com/v1/users

full = url.Join("https://example.com/base/", "path")!
puts full  # https://example.com/base/path
```

#### url.MustJoin(base, paths...) -> String

Joins paths, panicking on error.

```ruby
full = url.MustJoin("https://api.example.com", "v1", "users")
```

### Query String Encoding

#### url.QueryEncode(params) -> String

Encodes a map as a URL query string.

```ruby
query = url.QueryEncode({"q": "hello world", "page": "1"})
puts query  # page=1&q=hello+world
```

#### url.QueryEncodeValues(params) -> String

Encodes a map with multiple values per key.

```ruby
query = url.QueryEncodeValues({"tags": ["go", "ruby"]})
puts query  # tags=go&tags=ruby
```

#### url.QueryDecode(str) -> (Map[String, String], error)

Decodes a query string to a map (first value for each key).

```ruby
params = url.QueryDecode("a=1&b=2")!
puts params["a"]  # 1
```

#### url.QueryDecodeValues(str) -> (Map[String, Array[String]], error)

Decodes a query string with all values.

```ruby
params = url.QueryDecodeValues("a=1&a=2&b=3")!
puts params["a"]  # ["1", "2"]
```

### URL Encoding

#### url.Encode(str) -> String

URL-encodes a string (for query values).

```ruby
encoded = url.Encode("hello world")
puts encoded  # hello+world
```

#### url.Decode(str) -> (String, error)

URL-decodes a string.

```ruby
decoded = url.Decode("hello+world")!
puts decoded  # hello world
```

#### url.PathEncode(str) -> String

Encodes a string for use in a URL path.

```ruby
encoded = url.PathEncode("path with spaces")
puts encoded  # path%20with%20spaces
```

#### url.PathDecode(str) -> (String, error)

Decodes a URL path-encoded string.

```ruby
decoded = url.PathDecode("path%20with%20spaces")!
puts decoded  # path with spaces
```

### Validation

#### url.Valid(str) -> Bool

Reports whether a string can be parsed as a URL.

**Note:** Go's url.Parse is very permissive. Use `IsHTTP` for stricter validation.

```ruby
url.Valid("https://example.com")  # true
url.Valid("/path")                 # true
url.Valid("")                      # true
```

#### url.IsHTTP(str) -> Bool

Reports whether a URL has http or https scheme.

```ruby
url.IsHTTP("https://example.com")  # true
url.IsHTTP("http://localhost")     # true
url.IsHTTP("ftp://example.com")    # false
url.IsHTTP("/path")                # false
```

## Error Handling

```ruby
# Propagate error
u = url.Parse(input)!

# Provide default
u = url.Parse(input) rescue url.MustParse("https://default.com")

# Validate first
if url.IsHTTP(input)
  u = url.Parse(input)!
end
```

## Examples

### Build API URLs

```ruby
import rugby/url

class APIClient
  property base_url : String

  def endpoint(path : String, params : Map[String, String]) -> String
    u = url.MustParse(base_url)
    u = u.WithPath(path)
    if params.length > 0
      u = u.WithQuery(params)
    end
    u.String()
  end
end

client = APIClient{base_url: "https://api.example.com"}
endpoint = client.endpoint("/users", {"page": "1", "limit": "10"})
# https://api.example.com/users?limit=10&page=1
```

### Parse Query Parameters

```ruby
import rugby/url

def get_pagination(query : String) -> (Int, Int)
  params = url.QueryDecode(query) rescue {}
  page = params["page"].to_i() rescue 1
  limit = params["limit"].to_i() rescue 20
  page, limit
end

page, limit = get_pagination("page=2&limit=50")
```

### OAuth Redirect URL

```ruby
import rugby/url

def build_oauth_url(client_id : String, redirect : String, scopes : Array[String]) -> String
  params = {
    "client_id": client_id,
    "redirect_uri": redirect,
    "response_type": "code",
    "scope": scopes.join(" ")
  }
  base = url.MustParse("https://oauth.provider.com/authorize")
  base.WithQuery(params).String()
end
```

### URL-Safe Identifiers

```ruby
import rugby/url

def encode_path_segment(s : String) -> String
  url.PathEncode(s)
end

def decode_path_segment(s : String) -> String
  url.PathDecode(s) rescue s
end

# Safe for URL paths
safe = encode_path_segment("user name with spaces")
puts safe  # user%20name%20with%20spaces
```

### Extract Domain

```ruby
import rugby/url

def extract_domain(raw_url : String) -> String
  u = url.Parse(raw_url) rescue nil
  if u.nil?
    return ""
  end
  u.Hostname()
end

domain = extract_domain("https://www.example.com:8080/path")
puts domain  # www.example.com
```
