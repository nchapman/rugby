# rugby/base64

Base64 encoding and decoding for Rugby programs.

## Import

```ruby
import rugby/base64
```

## Quick Start

```ruby
import rugby/base64

# Encode a string
encoded = base64.EncodeString("Hello, World!")
puts encoded  # SGVsbG8sIFdvcmxkIQ==

# Decode back
decoded = base64.DecodeString(encoded)!
puts decoded  # Hello, World!

# Validate
if base64.Valid(encoded)
  puts "Valid base64"
end
```

## Functions

### Standard Encoding

#### base64.Encode(bytes) -> String

Encodes bytes to a base64 string.

```ruby
data = "hello".bytes
encoded = base64.Encode(data)  # aGVsbG8=
```

#### base64.EncodeString(str) -> String

Encodes a string to base64.

```ruby
encoded = base64.EncodeString("hello")  # aGVsbG8=
```

#### base64.Decode(str) -> (Bytes, error)

Decodes a base64 string to bytes.

```ruby
data = base64.Decode("aGVsbG8=")!
```

#### base64.DecodeString(str) -> (String, error)

Decodes a base64 string to a string.

```ruby
text = base64.DecodeString("aGVsbG8=")!  # hello
```

### URL-Safe Encoding

URL-safe encoding uses `-` instead of `+` and `_` instead of `/`, making it safe for URLs and filenames.

#### base64.URLEncode(bytes) -> String

Encodes bytes to URL-safe base64.

```ruby
encoded = base64.URLEncode(data)
```

#### base64.URLEncodeString(str) -> String

Encodes a string to URL-safe base64.

```ruby
encoded = base64.URLEncodeString("hello+world")
```

#### base64.URLDecode(str) -> (Bytes, error)

Decodes URL-safe base64 to bytes.

```ruby
data = base64.URLDecode(encoded)!
```

#### base64.URLDecodeString(str) -> (String, error)

Decodes URL-safe base64 to a string.

```ruby
text = base64.URLDecodeString(encoded)!
```

### Raw Encoding (No Padding)

Raw encoding omits the `=` padding characters.

#### base64.RawEncode(bytes) -> String

Encodes bytes to base64 without padding.

```ruby
encoded = base64.RawEncode(data)  # No trailing =
```

#### base64.RawEncodeString(str) -> String

Encodes a string to base64 without padding.

```ruby
encoded = base64.RawEncodeString("hello")  # aGVsbG8
```

#### base64.RawDecode(str) -> (Bytes, error)

Decodes unpadded base64 to bytes.

```ruby
data = base64.RawDecode("aGVsbG8")!
```

#### base64.RawDecodeString(str) -> (String, error)

Decodes unpadded base64 to a string.

```ruby
text = base64.RawDecodeString("aGVsbG8")!
```

### Raw URL-Safe Encoding

Combines URL-safe encoding with no padding.

#### base64.RawURLEncode(bytes) -> String

Encodes bytes to URL-safe base64 without padding.

```ruby
encoded = base64.RawURLEncode(data)
```

#### base64.RawURLEncodeString(str) -> String

Encodes a string to URL-safe base64 without padding.

```ruby
encoded = base64.RawURLEncodeString("hello")
```

#### base64.RawURLDecode(str) -> (Bytes, error)

Decodes URL-safe unpadded base64 to bytes.

```ruby
data = base64.RawURLDecode(encoded)!
```

#### base64.RawURLDecodeString(str) -> (String, error)

Decodes URL-safe unpadded base64 to a string.

```ruby
text = base64.RawURLDecodeString(encoded)!
```

### Validation

#### base64.Valid(str) -> Bool

Reports whether the string is valid standard base64.

```ruby
base64.Valid("aGVsbG8=")   # true
base64.Valid("!!!")        # false
base64.Valid("aGVsbG8")    # false (missing padding)
```

#### base64.URLValid(str) -> Bool

Reports whether the string is valid URL-safe base64.

```ruby
base64.URLValid("aGVsbG8=")  # true
base64.URLValid("-_-_")      # true (URL-safe chars)
```

#### base64.RawValid(str) -> Bool

Reports whether the string is valid unpadded base64.

```ruby
base64.RawValid("aGVsbG8")   # true (no padding required)
base64.RawValid("aGVsbG8=")  # false (padding not allowed)
```

#### base64.RawURLValid(str) -> Bool

Reports whether the string is valid unpadded URL-safe base64.

```ruby
base64.RawURLValid("aGVsbG8")  # true
```

## Error Handling

Decode functions return errors for invalid input:

```ruby
# Propagate error
decoded = base64.Decode(input)!

# Provide default
decoded = base64.Decode(input) rescue default_bytes

# Validate first
if base64.Valid(input)
  decoded = base64.Decode(input)!
end
```

## Examples

### Encode Binary Data

```ruby
import rugby/base64
import rugby/file

# Read binary file and encode
data = file.ReadBytes("image.png")!
encoded = base64.Encode(data)
puts "data:image/png;base64,#{encoded}"
```

### Basic Authentication

```ruby
import rugby/base64
import rugby/http

def basic_auth_header(user : String, pass : String) -> String
  credentials = "#{user}:#{pass}"
  encoded = base64.EncodeString(credentials)
  "Basic #{encoded}"
end

resp = http.Get(url, http.Options{
  Headers: {"Authorization": basic_auth_header("alice", "secret")}
})!
```

### JWT Token Parts

```ruby
import rugby/base64
import rugby/json

def decode_jwt_payload(token : String) -> Map[String, any]
  parts = token.split(".")
  if parts.length != 3
    panic "Invalid JWT format"
  end

  # JWT uses URL-safe base64 without padding
  payload = base64.RawURLDecodeString(parts[1])!
  json.Parse(payload)!
end

token = "eyJhbG..."
payload = decode_jwt_payload(token)
puts payload["sub"]
```

### Safe Filename Encoding

```ruby
import rugby/base64

def encode_filename(name : String) -> String
  # URL-safe, no padding = safe for filenames
  base64.RawURLEncodeString(name)
end

def decode_filename(encoded : String) -> String
  base64.RawURLDecodeString(encoded)!
end

safe = encode_filename("file/with:special*chars")
original = decode_filename(safe)
```
