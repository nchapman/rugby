# rugby/crypto

Cryptographic utilities for Rugby programs.

## Import

```ruby
import rugby/crypto
```

## Quick Start

```ruby
import rugby/crypto

# Hash a string
hash = crypto.SHA256String("hello world")
puts hash  # b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9

# Generate random bytes
token = crypto.RandomHex(16)!
puts token  # 32 random hex characters

# HMAC for authentication
sig = crypto.HMACSHA256String("secret-key", "message")
if crypto.VerifyHMACSHA256([]byte("secret-key"), []byte("message"), sig)
  puts "Valid signature"
end
```

## Functions

### Hash Functions

All hash functions come in three variants:
- `Hash(bytes)` - Takes bytes, returns hex string
- `HashString(str)` - Takes string, returns hex string
- `HashBytes(bytes)` - Takes bytes, returns raw bytes

#### SHA-256 (Recommended)

```ruby
hash = crypto.SHA256(data)           # bytes -> hex string
hash = crypto.SHA256String("hello")  # string -> hex string
bytes = crypto.SHA256Bytes(data)     # bytes -> raw bytes (32 bytes)
```

#### SHA-512

```ruby
hash = crypto.SHA512(data)           # bytes -> hex string
hash = crypto.SHA512String("hello")  # string -> hex string
bytes = crypto.SHA512Bytes(data)     # bytes -> raw bytes (64 bytes)
```

#### SHA-1 (Deprecated)

**Note:** SHA-1 is cryptographically broken. Use SHA-256 for security.

```ruby
hash = crypto.SHA1(data)             # bytes -> hex string
hash = crypto.SHA1String("hello")    # string -> hex string
bytes = crypto.SHA1Bytes(data)       # bytes -> raw bytes (20 bytes)
```

#### MD5 (Deprecated)

**Note:** MD5 is cryptographically broken. Use SHA-256 for security.

```ruby
hash = crypto.MD5(data)              # bytes -> hex string
hash = crypto.MD5String("hello")     # string -> hex string
bytes = crypto.MD5Bytes(data)        # bytes -> raw bytes (16 bytes)
```

### Random Generation

#### crypto.RandomBytes(n) -> (Bytes, error)

Generates n cryptographically secure random bytes.

```ruby
key = crypto.RandomBytes(32)!  # 256-bit key
iv = crypto.RandomBytes(16)!   # 128-bit IV
```

#### crypto.RandomHex(n) -> (String, error)

Generates n random bytes and returns them as a hex string.

```ruby
token = crypto.RandomHex(16)!  # 32-character hex string
session_id = crypto.RandomHex(32)!
```

### Constant-Time Comparison

Use these for comparing secrets to prevent timing attacks.

#### crypto.ConstantTimeCompare(a, b) -> Bool

Compares two byte slices in constant time.

```ruby
if crypto.ConstantTimeCompare(received, expected)
  puts "Match"
end
```

#### crypto.ConstantTimeCompareString(a, b) -> Bool

Compares two strings in constant time.

```ruby
if crypto.ConstantTimeCompareString(input_token, stored_token)
  puts "Valid token"
end
```

### HMAC (Message Authentication)

HMAC functions compute keyed hashes for message authentication.

#### HMAC-SHA256

```ruby
sig = crypto.HMACSHA256(key, data)           # bytes -> hex string
sig = crypto.HMACSHA256String("key", "msg")  # strings -> hex string
bytes = crypto.HMACSHA256Bytes(key, data)    # bytes -> raw bytes
```

#### HMAC-SHA512

```ruby
sig = crypto.HMACSHA512(key, data)           # bytes -> hex string
sig = crypto.HMACSHA512String("key", "msg")  # strings -> hex string
bytes = crypto.HMACSHA512Bytes(key, data)    # bytes -> raw bytes
```

### HMAC Verification

These functions verify signatures using constant-time comparison.

#### crypto.VerifyHMACSHA256(key, data, signature) -> Bool

Verifies a hex-encoded HMAC-SHA256 signature.

```ruby
sig = crypto.HMACSHA256(key, data)
if crypto.VerifyHMACSHA256(key, data, sig)
  puts "Valid"
end
```

#### crypto.VerifyHMACSHA256Bytes(key, data, signature) -> Bool

Verifies a raw byte HMAC-SHA256 signature.

```ruby
sig = crypto.HMACSHA256Bytes(key, data)
if crypto.VerifyHMACSHA256Bytes(key, data, sig)
  puts "Valid"
end
```

#### crypto.VerifyHMACSHA512(key, data, signature) -> Bool

Verifies a hex-encoded HMAC-SHA512 signature.

#### crypto.VerifyHMACSHA512Bytes(key, data, signature) -> Bool

Verifies a raw byte HMAC-SHA512 signature.

## Error Handling

Only `RandomBytes` and `RandomHex` can fail:

```ruby
# Propagate error
key = crypto.RandomBytes(32)!

# Handle error
key = crypto.RandomBytes(32) rescue default_key
```

## Examples

### Password Hashing (Simple)

For real password hashing, use bcrypt or argon2. This is a simple example:

```ruby
import rugby/crypto

def hash_password(password : String, salt : String) -> String
  crypto.SHA256String(salt + password)
end

def verify_password(password : String, salt : String, hash : String) -> Bool
  crypto.ConstantTimeCompareString(hash_password(password, salt), hash)
end
```

### API Request Signing

```ruby
import rugby/crypto
import rugby/time

def sign_request(secret : String, method : String, path : String) -> String
  timestamp = time.Now().Unix().to_s
  message = "#{method}\n#{path}\n#{timestamp}"
  crypto.HMACSHA256String(secret, message)
end

def verify_request(secret : String, method : String, path : String, signature : String) -> Bool
  expected = sign_request(secret, method, path)
  crypto.ConstantTimeCompareString(expected, signature)
end
```

### Generate Session Token

```ruby
import rugby/crypto

def generate_session_token -> String
  crypto.RandomHex(32)!  # 64-character hex string
end

def generate_api_key -> String
  crypto.RandomHex(24)!  # 48-character hex string
end
```

### Webhook Signature Verification

```ruby
import rugby/crypto

def verify_webhook(secret : String, payload : String, signature : String) -> Bool
  expected = "sha256=" + crypto.HMACSHA256String(secret, payload)
  crypto.ConstantTimeCompareString(expected, signature)
end

# Usage
if verify_webhook(webhook_secret, body, headers["X-Hub-Signature-256"])
  process_webhook(body)
end
```

### File Checksum

```ruby
import rugby/crypto
import rugby/file

def file_checksum(path : String) -> String
  data = file.ReadBytes(path)!
  crypto.SHA256(data)
end

def verify_file(path : String, expected_hash : String) -> Bool
  actual = file_checksum(path)
  crypto.ConstantTimeCompareString(actual, expected_hash)
end
```

### Content-Based Addressing

```ruby
import rugby/crypto

def content_address(data : Bytes) -> String
  crypto.SHA256(data)
end

def store_content(data : Bytes) -> String
  addr = content_address(data)
  file.WriteBytes("store/#{addr}", data)!
  addr
end
```
