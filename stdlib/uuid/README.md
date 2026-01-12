# rugby/uuid

UUID generation and parsing for Rugby programs.

## Import

```ruby
import rugby/uuid
```

## Quick Start

```ruby
import rugby/uuid

# Generate a random UUID
id = uuid.V4()!
puts id  # "550e8400-e29b-41d4-a716-446655440000"

# Generate a time-ordered UUID
id = uuid.V7()!
puts id  # UUIDs sorted by creation time

# Parse and validate
parsed = uuid.Parse("550E8400-E29B-41D4-A716-446655440000")!
puts parsed  # "550e8400-e29b-41d4-a716-446655440000" (normalized)

if uuid.Valid?("550e8400-e29b-41d4-a716-446655440000")
  puts "Valid UUID"
end
```

## Functions

### Generation

#### uuid.V4() -> (String, error)

Generates a random UUID (version 4) as a lowercase string.

```ruby
id = uuid.V4()!
puts id  # "550e8400-e29b-41d4-a716-446655440000"
```

#### uuid.V7() -> (String, error)

Generates a time-ordered UUID (version 7). V7 UUIDs are sortable by creation time and use Unix milliseconds.

```ruby
id1 = uuid.V7()!
sleep(0.01)
id2 = uuid.V7()!
# id1 < id2 (lexicographically)
```

**Use V7 for:**
- Database primary keys (sortable, indexable)
- Event IDs with temporal ordering
- Distributed ID generation without coordination

#### uuid.Deterministic(namespace, name) -> String

Generates a deterministic UUID from a namespace and name using SHA-256. Same inputs always produce the same UUID.

```ruby
id = uuid.Deterministic("users", "alice@example.com")
# Always returns the same UUID for this namespace+name
```

**Use for:**
- Content-addressable storage
- Idempotent operations
- Reproducible test data

#### uuid.MustV4() -> String

Like `V4()` but panics on error. Use when UUID generation failure is unrecoverable.

```ruby
id = uuid.MustV4()  # Panics on error
```

#### uuid.MustV7() -> String

Like `V7()` but panics on error.

```ruby
id = uuid.MustV7()  # Panics on error
```

### Validation

#### uuid.Parse(str) -> (String, error)

Parses a UUID string and returns the normalized lowercase form with hyphens. Accepts UUIDs with or without hyphens, in any case.

```ruby
# All of these parse to the same normalized form:
uuid.Parse("550e8400-e29b-41d4-a716-446655440000")!
uuid.Parse("550E8400-E29B-41D4-A716-446655440000")!
uuid.Parse("550e8400e29b41d4a716446655440000")!
# All return: "550e8400-e29b-41d4-a716-446655440000"
```

#### uuid.Valid?(str) -> Bool

Reports whether a string is a valid UUID. Accepts UUIDs with or without hyphens, in any case.

```ruby
if uuid.Valid?("550e8400-e29b-41d4-a716-446655440000")
  puts "Valid"
end

if uuid.Valid?("invalid")
  puts "This won't print"
end
```

### Inspection

#### uuid.Version(str) -> Int

Returns the UUID version number (1-7). Returns 0 if invalid.

```ruby
v4_id = uuid.V4()!
puts uuid.Version(v4_id)  # 4

v7_id = uuid.V7()!
puts uuid.Version(v7_id)  # 7
```

#### uuid.Timestamp(str) -> Int64

Extracts the timestamp from a V7 UUID as Unix milliseconds. Returns 0 for non-V7 UUIDs.

```ruby
id = uuid.V7()!
ms = uuid.Timestamp(id)
puts ms  # Unix milliseconds
```

#### uuid.Time(str) -> time.Time

Extracts the timestamp from a V7 UUID as a time.Time. Returns zero time for non-V7 UUIDs.

```ruby
import rugby/time

id = uuid.V7()!
t = uuid.Time(id)
puts time.Format(t, time.RFC3339)
```

### Comparison

#### uuid.Compare(a, b) -> Int

Compares two UUIDs lexicographically. Returns -1 if a < b, 0 if equal, 1 if a > b. Returns 0 if either is invalid.

```ruby
a = "00000000-0000-0000-0000-000000000001"
b = "00000000-0000-0000-0000-000000000002"
puts uuid.Compare(a, b)  # -1
```

#### uuid.Equal?(a, b) -> Bool

Reports whether two UUIDs are equal. Handles case differences and hyphen presence/absence.

```ruby
a = "550e8400-e29b-41d4-a716-446655440000"
b = "550E8400E29B41D4A716446655440000"
if uuid.Equal?(a, b)
  puts "Same UUID"
end
```

### Nil UUID

#### uuid.Nil() -> String

Returns the nil UUID (all zeros).

```ruby
puts uuid.Nil()  # "00000000-0000-0000-0000-000000000000"
```

#### uuid.Nil?(str) -> Bool

Reports whether a string is the nil UUID.

```ruby
if uuid.Nil?("00000000-0000-0000-0000-000000000000")
  puts "This is the nil UUID"
end
```

### Binary Conversion

#### uuid.Bytes(str) -> (Bytes, error)

Converts a UUID string to its 16-byte representation.

```ruby
bytes = uuid.Bytes("550e8400-e29b-41d4-a716-446655440000")!
puts bytes.length  # 16
```

#### uuid.FromBytes(bytes) -> (String, error)

Creates a UUID string from a 16-byte slice.

```ruby
bytes = [0x55, 0x0e, 0x84, 0x00, ...]  # 16 bytes
id = uuid.FromBytes(bytes)!
puts id  # "550e8400-e29b-41d4-a716-446655440000"
```

## Error Handling

UUID generation can fail (rarely) due to random number generator errors:

```ruby
# Propagate error
id = uuid.V4()!

# Handle error
id = uuid.V4() rescue uuid.Nil()

# Use if let
if let id = uuid.V4()
  puts "Generated: #{id}"
end
```

Parsing errors occur with invalid input:

```ruby
# Propagate error
id = uuid.Parse(input)!

# Handle invalid input
id = uuid.Parse(input) rescue do
  puts "Invalid UUID"
  uuid.Nil()
end
```

## Examples

### Database IDs

Use V7 for time-ordered database primary keys:

```ruby
import rugby/uuid

class User
  def initialize(@id : String, @name : String)
  end
end

def create_user(name : String) -> User
  id = uuid.V7()!
  User.new(id, name)
end

# IDs are naturally sorted by creation time
users = [create_user("Alice"), create_user("Bob"), create_user("Charlie")]
# users.map { |u| u.id } is sorted chronologically
```

### Idempotent Operations

Use deterministic UUIDs for idempotent operations:

```ruby
import rugby/uuid

def process_email(email : String)
  # Same email always gets same ID
  id = uuid.Deterministic("email-jobs", email)

  if job_exists?(id)
    puts "Already processed"
    return
  end

  create_job(id, email)
  send_email(email)
end
```

### Request Tracing

Use V4 for request IDs:

```ruby
import rugby/uuid
import rugby/log

def handle_request(req)
  request_id = uuid.V4()!
  logger = log.With({"request_id": request_id})

  logger.Info("request started")
  result = process_request(req)
  logger.Info("request completed")

  result
end
```

### Event Sourcing

Use V7 for event IDs to maintain temporal ordering:

```ruby
import rugby/uuid
import rugby/time

class Event
  property id : String
  property timestamp : Int64
  property type : String
  property data : Map[String, any]
end

def create_event(type : String, data : Map[String, any]) -> Event
  id = uuid.V7()!
  Event.new(
    id: id,
    timestamp: uuid.Timestamp(id),
    type: type,
    data: data
  )
end

# Events are naturally sorted by ID
events = [
  create_event("user.created", {"user_id": 1}),
  create_event("user.updated", {"user_id": 1}),
  create_event("user.deleted", {"user_id": 1})
]
# events.map { |e| e.id } is sorted chronologically
```

### URL Shortener

Use deterministic UUIDs for URL shortening:

```ruby
import rugby/uuid
import rugby/base64

def shorten_url(url : String) -> String
  id = uuid.Deterministic("urls", url)
  bytes = uuid.Bytes(id)!
  # Take first 8 bytes for shorter ID
  base64.RawURLEncodeString(bytes[0...8].to_s)
end

puts shorten_url("https://example.com")  # Always same short ID
```

### File Deduplication

Use content-based UUIDs for deduplication:

```ruby
import rugby/uuid
import rugby/crypto

def store_file(data : Bytes) -> String
  # Hash content to get deterministic ID
  hash = crypto.SHA256(data)
  id = uuid.Deterministic("files", hash)

  if file_exists?(id)
    puts "File already exists"
    return id
  end

  write_file(id, data)
  id
end
```

### Testing

Use deterministic UUIDs for reproducible tests:

```ruby
import rugby/uuid

def test_user_creation
  # Same UUID every test run
  test_id = uuid.Deterministic("test", "user-1")

  user = create_user(test_id, "Alice")
  assert user.id == test_id
end
```

## Performance

- V4 generation: ~1 microsecond
- V7 generation: ~1 microsecond
- Parse/validation: ~100 nanoseconds
- No allocations for validation

V7 UUIDs include a timestamp prefix, making them ideal for database indexes.
