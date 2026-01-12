# rugby/env

Environment variable access for Rugby programs.

## Import

```ruby
import rugby/env
```

## Quick Start

```ruby
import rugby/env

# Get with default
port = env.Fetch("PORT", "8080")
puts "Starting on port #{port}"

# Check if variable exists
if let home = env.Get("HOME")
  puts "Home directory: #{home}"
end

# Set a variable
env.Set("APP_MODE", "production")!
```

## Functions

### Reading

#### env.Get(key) -> (String, Bool)

Returns the value of an environment variable and whether it was set.

```ruby
value, ok = env.Get("API_KEY")
if ok
  puts "API key is set"
else
  puts "API key not found"
end

# Or use if let
if let value = env.Get("API_KEY")
  use_api_key(value)
end
```

#### env.Fetch(key, default) -> String

Returns the value of an environment variable, or the default if not set.

```ruby
port = env.Fetch("PORT", "3000")
host = env.Fetch("HOST", "localhost")
debug = env.Fetch("DEBUG", "false")
```

#### env.Has(key) -> Bool

Reports whether an environment variable is set.

```ruby
if env.Has("PRODUCTION")
  # use production settings
end
```

### Writing

#### env.Set(key, value) -> error

Sets an environment variable.

```ruby
env.Set("APP_NAME", "MyApp")!
```

#### env.Delete(key) -> error

Removes an environment variable.

```ruby
env.Delete("TEMP_VAR")!
```

#### env.Clear()

Removes all environment variables. Use with caution!

```ruby
env.Clear()  # Clears everything
```

### Listing

#### env.All() -> Map[String, String]

Returns all environment variables as a map.

```ruby
all_vars = env.All()
for key, value in all_vars
  puts "#{key}=#{value}"
end
```

#### env.Keys() -> Array[String]

Returns all environment variable names.

```ruby
keys = env.Keys()
puts "#{keys.length} environment variables set"
```

## Error Handling

`Set` and `Delete` return errors, though they rarely fail:

```ruby
env.Set("KEY", "value")!

# Or handle explicitly
err = env.Set("KEY", "value")
if err != nil
  puts "Failed to set: #{err}"
end
```

## Examples

### Configuration from Environment

```ruby
import rugby/env

class Config
  property port : Int
  property host : String
  property debug : Bool
end

def load_config -> Config
  Config{
    port: env.Fetch("PORT", "8080").to_i() rescue 8080,
    host: env.Fetch("HOST", "0.0.0.0"),
    debug: env.Fetch("DEBUG", "false") == "true"
  }
end

config = load_config()
puts "Server: #{config.host}:#{config.port}"
```

### Required Variables

```ruby
import rugby/env

def require_env(key : String) -> String
  value, ok = env.Get(key)
  unless ok
    panic "Required environment variable #{key} not set"
  end
  value
end

api_key = require_env("API_KEY")
db_url = require_env("DATABASE_URL")
```

### Debug Environment

```ruby
import rugby/env

def print_env
  puts "Environment Variables:"
  puts "====================="
  all = env.All()
  for key, value in all
    # Mask sensitive values
    display = if key.contains?("KEY") || key.contains?("SECRET")
      "****"
    else
      value
    end
    puts "#{key}=#{display}"
  end
end
```

### Temporary Environment

```ruby
import rugby/env

def with_env(key : String, value : String, block : -> any) -> any
  old, had_old = env.Get(key)
  env.Set(key, value)!

  result = block()

  if had_old
    env.Set(key, old)!
  else
    env.Delete(key)!
  end

  result
end

# Usage
with_env("DEBUG", "true") do
  run_tests()
end
```
