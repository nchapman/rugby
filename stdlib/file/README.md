# rugby/file

File I/O operations for Rugby programs.

## Import

```ruby
import rugby/file
```

## Quick Start

```ruby
import rugby/file

# Read a file
content = file.Read("config.txt")!
puts content

# Write a file
file.Write("output.txt", "Hello, world!")!

# Check if file exists
if file.Exists("data.json")
  data = file.Read("data.json")!
end
```

## Functions

### Reading

#### file.Read(path) -> (String, error)

Reads a file and returns its contents as a string.

```ruby
content = file.Read("readme.txt")!
```

#### file.ReadBytes(path) -> (Bytes, error)

Reads a file and returns its contents as bytes.

```ruby
data = file.ReadBytes("image.png")!
```

#### file.ReadLines(path) -> (Array[String], error)

Reads a file and returns its lines as an array.

```ruby
lines = file.ReadLines("data.csv")!
for line in lines
  puts line
end
```

### Writing

#### file.Write(path, content) -> error

Writes a string to a file, creating it if it doesn't exist.

```ruby
file.Write("output.txt", "Hello, world!")!
```

#### file.WriteBytes(path, data) -> error

Writes bytes to a file.

```ruby
file.WriteBytes("output.bin", binary_data)!
```

#### file.WriteLines(path, lines) -> error

Writes an array of strings to a file, joining them with newlines.

```ruby
file.WriteLines("list.txt", ["apple", "banana", "cherry"])!
```

#### file.Append(path, content) -> error

Appends a string to a file, creating it if it doesn't exist.

```ruby
file.Append("log.txt", "New entry\n")!
```

#### file.AppendBytes(path, data) -> error

Appends bytes to a file.

```ruby
file.AppendBytes("data.bin", more_bytes)!
```

### File Info

#### file.Exists(path) -> Bool

Reports whether a file or directory exists.

```ruby
if file.Exists("config.json")
  # load config
end
```

#### file.Size(path) -> (Int64, error)

Returns the size of a file in bytes.

```ruby
size = file.Size("data.bin")!
puts "File is #{size} bytes"
```

#### file.Directory(path) -> Bool

Reports whether the path is a directory.

```ruby
if file.Directory("src")
  puts "src is a directory"
end
```

#### file.File(path) -> Bool

Reports whether the path is a regular file.

```ruby
if file.File("main.rg")
  puts "main.rg is a file"
end
```

### File Operations

#### file.Remove(path) -> error

Deletes a file or empty directory.

```ruby
file.Remove("temp.txt")!
```

#### file.RemoveAll(path) -> error

Deletes a path and all its contents recursively.

```ruby
file.RemoveAll("build/")!
```

#### file.Rename(old, new) -> error

Renames (moves) a file or directory.

```ruby
file.Rename("old_name.txt", "new_name.txt")!
```

#### file.Copy(src, dst) -> error

Copies a file, preserving permissions.

```ruby
file.Copy("original.txt", "backup.txt")!
```

### Directory Operations

#### file.Mkdir(path) -> error

Creates a directory.

```ruby
file.Mkdir("output")!
```

#### file.MkdirAll(path) -> error

Creates a directory and all parent directories.

```ruby
file.MkdirAll("path/to/deep/dir")!
```

## Error Handling

All functions that can fail return errors. Use `!` to propagate or `rescue` for fallbacks:

```ruby
# Propagate error
content = file.Read("config.txt")!

# Provide default
content = file.Read("config.txt") rescue "{}"

# Handle explicitly
content, err = file.Read("config.txt")
if err != nil
  puts "Could not read config: #{err}"
end
```

## Examples

### Read and Process Config

```ruby
import rugby/file
import rugby/json

content = file.Read("config.json")!
config = json.Parse(content)!
puts "Port: #{config["port"]}"
```

### Log to File

```ruby
import rugby/file
import rugby/time

def log(message : String)
  timestamp = time.Now().Format("2006-01-02 15:04:05")
  file.Append("app.log", "[#{timestamp}] #{message}\n")!
end

log("Application started")
```

### Safe File Write

```ruby
import rugby/file

def safe_write(path : String, content : String) -> error
  temp = path + ".tmp"
  file.Write(temp, content)!
  file.Rename(temp, path)!
  nil
end
```

### Copy Directory Contents

```ruby
import rugby/file
import rugby/path

def copy_file(src : String, dst : String) -> error
  if file.Directory(src)
    return file.MkdirAll(dst)
  end
  file.Copy(src, dst)
end
```
