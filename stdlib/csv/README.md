# rugby/csv

CSV parsing and generation for Rugby programs.

## Import

```ruby
import rugby/csv
```

## Quick Start

```ruby
import rugby/csv

# Parse CSV
rows = csv.Parse("a,b,c\n1,2,3")!
puts rows[0]  # ["a", "b", "c"]
puts rows[1]  # ["1", "2", "3"]

# Parse with headers
data = csv.ParseWithHeaders("name,age\nAlice,30\nBob,25")!
puts data[0]["name"]  # "Alice"
puts data[0]["age"]   # "30"

# Generate CSV
output = csv.Generate([
  ["name", "age"],
  ["Alice", "30"],
  ["Bob", "25"]
])!
puts output  # "name,age\nAlice,30\nBob,25\n"
```

## Functions

### Parsing

#### csv.Parse(str) -> (Array[Array[String]], error)

Parses a CSV string into rows of fields.

```ruby
rows = csv.Parse("a,b,c\n1,2,3")!
# [["a", "b", "c"], ["1", "2", "3"]]
```

Handles quoted fields and escaped quotes:

```ruby
rows = csv.Parse('"hello, world",b,c')!
# [["hello, world", "b", "c"]]

rows = csv.Parse('"say ""hello""",b')!
# [["say \"hello\"", "b"]]
```

#### csv.ParseBytes(bytes) -> (Array[Array[String]], error)

Parses CSV bytes into rows of fields.

```ruby
rows = csv.ParseBytes(data)!
```

#### csv.ParseWithHeaders(str) -> (Array[Map[String, String]], error)

Parses a CSV string using the first row as headers. Returns a slice of maps where each map represents a row with header keys.

```ruby
data = csv.ParseWithHeaders("name,age\nAlice,30\nBob,25")!
# [
#   {"name": "Alice", "age": "30"},
#   {"name": "Bob", "age": "25"}
# ]

puts data[0]["name"]  # "Alice"
```

If a row has fewer fields than headers, missing fields are empty strings:

```ruby
data = csv.ParseWithHeaders("name,age,city\nAlice,30")!
# [{"name": "Alice", "age": "30", "city": ""}]
```

#### csv.ParseBytesWithHeaders(bytes) -> (Array[Map[String, String]], error)

Parses CSV bytes using the first row as headers.

```ruby
data = csv.ParseBytesWithHeaders(bytes)!
```

### Generation

#### csv.Generate(rows) -> (String, error)

Converts rows of fields to a CSV string.

```ruby
output = csv.Generate([
  ["name", "age"],
  ["Alice", "30"],
  ["Bob", "25"]
])!
# "name,age\nAlice,30\nBob,25\n"
```

Automatically quotes fields that contain commas, quotes, or newlines:

```ruby
output = csv.Generate([["hello, world", "b"]])!
# "\"hello, world\",b\n"
```

#### csv.GenerateBytes(rows) -> (Bytes, error)

Converts rows of fields to CSV bytes.

```ruby
bytes = csv.GenerateBytes(rows)!
```

#### csv.GenerateWithHeaders(headers, records) -> (String, error)

Generates a CSV string from maps using specified headers. The headers define the order of columns in the output.

```ruby
headers = ["name", "age"]
records = [
  {"name": "Alice", "age": "30"},
  {"name": "Bob", "age": "25"}
]

output = csv.GenerateWithHeaders(headers, records)!
# "name,age\nAlice,30\nBob,25\n"
```

### Streaming

For large files, use streaming readers and writers.

#### csv.Reader(str) -> Reader

Creates a reader from a string.

```ruby
reader = csv.Reader(data)

for row = reader.Read(); row; row = reader.Read()
  puts row[0]
end
```

Or use `ReadAll()`:

```ruby
reader = csv.Reader(data)
rows = reader.ReadAll()!
```

**Reader methods:**
- `Read() -> (Array[String], error)` - Read next row
- `ReadAll() -> (Array[Array[String]], error)` - Read all remaining rows
- `SetDelimiter(delim)` - Set field delimiter (default: comma)
- `SetComment(c)` - Set comment character (lines starting with this are ignored)
- `SetLazyQuotes(lazy)` - Allow lazy handling of quotes
- `SetTrimLeadingSpace(trim)` - Trim leading whitespace from fields

#### csv.ReaderBytes(bytes) -> Reader

Creates a reader from bytes.

```ruby
reader = csv.ReaderBytes(data)
```

#### csv.Writer() -> Writer

Creates a new CSV writer.

```ruby
writer = csv.Writer()
writer.Write(["name", "age"])!
writer.Write(["Alice", "30"])!

output = writer.String()
# "name,age\nAlice,30\n"
```

**Writer methods:**
- `Write(row) -> error` - Write a single row
- `WriteAll(rows) -> error` - Write multiple rows
- `SetDelimiter(delim)` - Set field delimiter (default: comma)
- `SetUseCRLF(use)` - Use \r\n as line terminator (for Windows)
- `Flush()` - Flush buffered data
- `String() -> String` - Get CSV data as string
- `Bytes() -> Bytes` - Get CSV data as bytes
- `Error() -> error` - Get any error that occurred

### Utilities

#### csv.ParseLine(str) -> (Array[String], error)

Parses a single CSV line into fields.

```ruby
fields = csv.ParseLine("a,b,c")!
# ["a", "b", "c"]
```

#### csv.Escape(str) -> String

Quotes a field for CSV output if needed.

```ruby
puts csv.Escape("simple")      # "simple"
puts csv.Escape("hello, world")  # "\"hello, world\""
```

#### csv.Valid?(str) -> Bool

Reports whether a string is valid CSV.

```ruby
if csv.Valid?("a,b,c\n1,2,3")
  puts "Valid CSV"
end
```

#### csv.RowCount(str) -> Int

Returns the number of rows in a CSV string. Returns -1 if invalid.

```ruby
count = csv.RowCount("a,b,c\n1,2,3")
puts count  # 2
```

#### csv.Headers(str) -> (Array[String], error)

Returns the first row of a CSV as headers.

```ruby
headers = csv.Headers("name,age\nAlice,30")!
# ["name", "age"]
```

## Error Handling

CSV parsing can fail with invalid input:

```ruby
# Propagate error
rows = csv.Parse(input)!

# Handle error
rows = csv.Parse(input) rescue do
  puts "Invalid CSV"
  []
end

# Use if let
if let rows = csv.Parse(input)
  process_rows(rows)
end
```

## Examples

### Read CSV File

```ruby
import rugby/csv
import rugby/file

def load_csv(path : String) -> Array[Array[String]]
  content = file.Read(path)!
  csv.Parse(content)!
end

rows = load_csv("data.csv")
for row in rows
  puts row.join(", ")
end
```

### Parse with Headers

```ruby
import rugby/csv
import rugby/file

def load_users(path : String) -> Array[Map[String, String]]
  content = file.Read(path)!
  csv.ParseWithHeaders(content)!
end

users = load_users("users.csv")
for user in users
  puts "Name: #{user["name"]}, Age: #{user["age"]}"
end
```

### Generate and Save

```ruby
import rugby/csv
import rugby/file

def export_users(users, path : String)
  rows = users.map { |u| [u.name, u.age.to_s, u.email] }
  rows.unshift(["name", "age", "email"])  # Add header

  output = csv.Generate(rows)!
  file.Write(path, output)!
end
```

### Custom Delimiter

Use semicolons or tabs:

```ruby
import rugby/csv

reader = csv.Reader(data)
reader.SetDelimiter(';')
rows = reader.ReadAll()!

# Or for writing:
writer = csv.Writer()
writer.SetDelimiter('\t')
writer.Write(["a", "b", "c"])!
puts writer.String()  # "a\tb\tc\n"
```

### Comments

Ignore lines starting with `#`:

```ruby
import rugby/csv

data = "# This is a comment\na,b,c\n1,2,3"

reader = csv.Reader(data)
reader.SetComment('#')
rows = reader.ReadAll()!
# Only returns [["a", "b", "c"], ["1", "2", "3"]]
```

### Streaming Large Files

For large files, stream row by row:

```ruby
import rugby/csv
import rugby/file

def process_large_csv(path : String)
  content = file.Read(path)!
  reader = csv.Reader(content)

  # Skip header
  reader.Read()!

  count = 0
  while let row = reader.Read()
    process_row(row)
    count += 1
  end

  puts "Processed #{count} rows"
end
```

### Export with Headers

```ruby
import rugby/csv

class User
  property name : String
  property age : Int
  property email : String
end

def export_users(users : Array[User]) -> String
  headers = ["name", "age", "email"]
  records = users.map { |u|
    {"name": u.name, "age": u.age.to_s, "email": u.email}
  }

  csv.GenerateWithHeaders(headers, records)!
end
```

### Convert to JSON

```ruby
import rugby/csv
import rugby/json

def csv_to_json(csv_str : String) -> String
  data = csv.ParseWithHeaders(csv_str)!
  json.Pretty(data)!
end

csv_data = "name,age\nAlice,30\nBob,25"
json_output = csv_to_json(csv_data)
puts json_output
```

### Validate CSV

```ruby
import rugby/csv

def validate_csv(str : String) -> Bool
  return false unless csv.Valid?(str)

  rows = csv.Parse(str)!
  return false if rows.empty?

  # Check all rows have same number of columns
  columns = rows[0].length
  rows.all? { |row| row.length == columns }
end
```

### Filter and Transform

```ruby
import rugby/csv

def filter_by_age(csv_str : String, min_age : Int) -> String
  data = csv.ParseWithHeaders(csv_str)!

  filtered = data.select { |row|
    row["age"].to_i >= min_age
  }

  headers = ["name", "age"]
  csv.GenerateWithHeaders(headers, filtered)!
end

input = "name,age\nAlice,30\nBob,15\nCharlie,25"
output = filter_by_age(input, 18)
# Only Alice and Charlie
```

### Windows Line Endings

```ruby
import rugby/csv

writer = csv.Writer()
writer.SetUseCRLF(true)
writer.Write(["a", "b"])!

puts writer.String()  # "a,b\r\n"
```

### Trim Whitespace

```ruby
import rugby/csv

data = "  a ,  b  ,  c  "

reader = csv.Reader(data)
reader.SetTrimLeadingSpace(true)
row = reader.Read()!
# ["a", "b", "c"] (trimmed)
```

## Performance

- Parsing: ~100 MB/s
- Generation: ~150 MB/s
- Streaming: Constant memory regardless of file size
- Zero-copy parsing where possible

Use streaming readers for files larger than 100 MB to minimize memory usage.
