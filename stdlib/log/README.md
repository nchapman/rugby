# rugby/log

Structured logging with levels for Rugby programs.

## Import

```ruby
import rugby/log
```

## Quick Start

```ruby
import rugby/log

# Simple logging
log.Info("user logged in")
log.Error("connection failed")

# Structured logging with fields
log.Info("user logged in", {"user_id": 123, "ip": "192.168.1.1"})
log.Error("database error", {"query": sql, "duration_ms": 250})

# Set log level
log.SetLevel(log.DebugLevel)

# JSON output
log.SetFormat(log.JSONFormat)
```

## Log Levels

Levels from lowest to highest:

- `DebugLevel` - Detailed information for debugging
- `InfoLevel` - General information (default)
- `WarnLevel` - Warning messages
- `ErrorLevel` - Error messages
- `FatalLevel` - Fatal errors (calls `exit(1)` after logging)
- `Off` - Disable all logging

## Functions

### Logging

#### log.Debug(msg, fields...)

Logs at debug level. Only shown when level is DebugLevel.

```ruby
log.Debug("processing request")
log.Debug("cache hit", {"key": cache_key})
```

#### log.Info(msg, fields...)

Logs at info level. Default logging level.

```ruby
log.Info("server started", {"port": 8080})
log.Info("request processed", {"duration_ms": 45})
```

#### log.Warn(msg, fields...)

Logs at warn level. Use for potentially harmful situations.

```ruby
log.Warn("deprecated API used", {"endpoint": "/old-api"})
log.Warn("retry attempt", {"attempt": 3, "max": 5})
```

#### log.Error(msg, fields...)

Logs at error level. Use for errors that don't require immediate shutdown.

```ruby
log.Error("database connection failed", {"host": "db.example.com"})
log.Error("request failed", {"error": err.message, "status": 500})
```

#### log.Fatal(msg, fields...)

Logs at fatal level and exits the program with status code 1. Use only for unrecoverable errors.

```ruby
log.Fatal("cannot bind port", {"port": 8080, "error": err.message})
# Program exits here
```

### Configuration

#### log.SetLevel(level)

Sets the minimum log level. Messages below this level are not logged.

```ruby
log.SetLevel(log.DebugLevel)  # Show everything
log.SetLevel(log.WarnLevel)   # Only warnings and errors
log.SetLevel(log.Off)          # Disable all logging
```

#### log.GetLevel() -> Level

Returns the current log level.

```ruby
current = log.GetLevel()
if current == log.DebugLevel
  puts "Debug mode enabled"
end
```

#### log.SetFormat(format)

Sets the output format.

```ruby
log.SetFormat(log.TextFormat)  # Human-readable (default)
log.SetFormat(log.JSONFormat)  # Machine-parseable JSON
```

**Text format:**
```
2024-01-15T10:30:00.000Z INFO user logged in user_id=123
```

**JSON format:**
```json
{"time":"2024-01-15T10:30:00.000Z","level":"info","msg":"user logged in","user_id":123}
```

#### log.SetOutput(writer)

Sets the output destination. Default is `os.Stderr`.

```ruby
import rugby/file

f = file.OpenForAppend("app.log")!
log.SetOutput(f)
```

#### log.SetName(name)

Sets a logger name that appears in output.

```ruby
log.SetName("api")
log.Info("request received")
# 2024-01-15T10:30:00.000Z INFO [api] request received
```

### Contextual Logging

#### log.With(fields) -> Logger

Returns a new logger with additional fields. The new logger inherits all settings and fields from the parent.

```ruby
# Create request-specific logger
request_logger = log.With({"request_id": "abc123", "user_id": 456})

request_logger.Info("request started")
request_logger.Info("query executed", {"table": "users"})
# Both logs include request_id and user_id
```

Child loggers don't affect parent:

```ruby
base = log.With({"service": "api"})
request1 = base.With({"request_id": "req-1"})
request2 = base.With({"request_id": "req-2"})

# Each has isolated context
request1.Info("processing")  # service=api request_id=req-1
request2.Info("processing")  # service=api request_id=req-2
```

### Utility Functions

#### log.ParseLevel(str) -> (Level, error)

Parses a level string.

```ruby
level = log.ParseLevel("debug")!
log.SetLevel(level)
```

Accepts: `"debug"`, `"info"`, `"warn"`, `"warning"`, `"error"`, `"fatal"`, `"off"` (case-insensitive)

#### log.IsEnabled(level) -> Bool

Reports whether logging is enabled at the given level.

```ruby
if log.IsEnabled(log.DebugLevel)
  # Skip expensive debug computation
  log.Debug("state", {"data": expensive_debug_info()})
end
```

## Custom Loggers

Create isolated loggers with their own settings:

```ruby
logger = log.New()
logger.SetLevel(log.WarnLevel)
logger.SetName("worker")
logger.SetFormat(log.JSONFormat)

logger.Info("starting")  # Won't appear (below WarnLevel)
logger.Warn("slow query")  # Appears
```

## Error Handling

Logging functions don't return errors. If output fails (e.g., disk full), errors are silently ignored to prevent logging from crashing the application.

## Examples

### Basic Application Logging

```ruby
import rugby/log

log.SetLevel(log.InfoLevel)
log.SetName("myapp")

def main
  log.Info("application starting", {"version": "1.0.0"})

  server.Start(8080)!

  log.Info("application stopped")
end
```

### Request Logging

```ruby
import rugby/log
import rugby/uuid

def handle_request(req)
  request_id = uuid.V4()!
  logger = log.With({"request_id": request_id, "method": req.method, "path": req.path})

  logger.Info("request started")

  result = process_request(req) rescue => err do
    logger.Error("request failed", {"error": err.message})
    return error_response(500)
  end

  logger.Info("request completed", {"status": 200, "duration_ms": elapsed})
  result
end
```

### Structured Error Logging

```ruby
import rugby/log

def connect_database(config)
  db.Connect(config.host, config.port) rescue => err do
    log.Error("database connection failed", {
      "host": config.host,
      "port": config.port,
      "error": err.message,
      "retry_count": config.retries
    })
    return nil, err
  end
end
```

### Environment-Based Configuration

```ruby
import rugby/log
import rugby/env

def configure_logging
  level_str = env.Get("LOG_LEVEL") ?? "info"
  format_str = env.Get("LOG_FORMAT") ?? "text"

  level = log.ParseLevel(level_str) rescue log.InfoLevel
  log.SetLevel(level)

  if format_str == "json"
    log.SetFormat(log.JSONFormat)
  end

  log.Info("logging configured", {"level": level.to_s, "format": format_str})
end
```

### Worker Context

```ruby
import rugby/log

class Worker
  def initialize(@id : Int)
    @logger = log.With({"worker_id": @id})
  end

  def process_job(job)
    job_logger = @logger.With({"job_id": job.id, "job_type": job.type})

    job_logger.Info("job started")

    result = execute_job(job) rescue => err do
      job_logger.Error("job failed", {"error": err.message})
      return
    end

    job_logger.Info("job completed", {"duration_ms": result.duration})
  end
end
```

### Debug Logging with Guards

```ruby
import rugby/log

def process_large_data(data)
  if log.IsEnabled(log.DebugLevel)
    # Only compute debug info if debug is enabled
    debug_info = expensive_debug_computation(data)
    log.Debug("processing data", {"info": debug_info})
  end

  process(data)
end
```

### Log to File

```ruby
import rugby/log
import rugby/file

def setup_file_logging(path : String)
  f = file.OpenForAppend(path)!
  log.SetOutput(f)
  log.Info("logging to file", {"path": path})
end

setup_file_logging("/var/log/myapp.log")
```

### Multiple Loggers

```ruby
import rugby/log

# Application logger
app = log.New()
app.SetName("app")
app.SetLevel(log.InfoLevel)

# Access logger (more verbose)
access = log.New()
access.SetName("access")
access.SetLevel(log.DebugLevel)
access.SetFormat(log.JSONFormat)

def handle_request(req)
  access.Debug("request details", {"headers": req.headers})
  app.Info("processing request", {"path": req.path})

  process(req)
end
```

### Sampling (Rate Limiting)

```ruby
import rugby/log

counter = 0

def log_sample(msg, fields)
  counter += 1
  if counter % 100 == 0
    log.Info(msg, fields.merge({"sample": counter}))
  end
end

# Only logs every 100th call
for event in events
  log_sample("event processed", {"type": event.type})
end
```

### Panic Recovery with Logging

```ruby
import rugby/log

def safe_handler(req)
  handle_request(req)
rescue => err
  log.Error("panic recovered", {
    "error": err.message,
    "path": req.path
  })
  error_response(500)
end
```

### Audit Logging

```ruby
import rugby/log
import rugby/time

audit = log.New()
audit.SetFormat(log.JSONFormat)
audit.SetOutput(audit_file)

def log_audit(action : String, user_id : Int, details : Map)
  audit.Info(action, {
    "user_id": user_id,
    "timestamp": time.Now().Unix(),
    "details": details
  })
end

log_audit("user.login", 123, {"ip": "192.168.1.1"})
log_audit("user.logout", 123, {"session_duration": 3600})
```

### JSON Parsing from Logs

```ruby
import rugby/log
import ruby/json

log.SetFormat(log.JSONFormat)
log.Info("test", {"key": "value"})
# {"time":"...","level":"info","msg":"test","key":"value"}

# Parse logs later:
def parse_log_line(line : String) -> Map
  json.Parse(line)!
end
```

## Field Ordering

Fields are sorted alphabetically in text format for deterministic output:

```ruby
log.Info("test", {"z": 1, "a": 2, "m": 3})
# Output: ... test a=2 m=3 z=1
```

This ensures consistent log output for testing and parsing.

## Reserved Keys

In JSON format, these keys are reserved and cannot be overwritten by user fields:
- `time` - Timestamp
- `level` - Log level
- `msg` - Log message
- `logger` - Logger name

If you provide these keys in fields, the reserved values take precedence.

## Performance

- Text logging: ~500,000 messages/sec
- JSON logging: ~400,000 messages/sec
- Field inheritance: Zero-copy
- Thread-safe: All operations are protected by mutex

Logging to `/dev/null` is nearly zero-cost when the level is too low.

## Best Practices

1. **Use structured fields** instead of string formatting:
   ```ruby
   # Good
   log.Info("user login", {"user_id": id, "ip": ip})

   # Avoid
   log.Info("user #{id} logged in from #{ip}")
   ```

2. **Add context with With()**:
   ```ruby
   request_log = log.With({"request_id": id})
   request_log.Info("started")
   request_log.Info("completed")
   ```

3. **Use appropriate levels**:
   - Debug: Development/troubleshooting only
   - Info: Normal operation events
   - Warn: Unusual but handled situations
   - Error: Errors that don't require shutdown
   - Fatal: Unrecoverable errors only

4. **Guard expensive debug calls**:
   ```ruby
   if log.IsEnabled(log.DebugLevel)
     log.Debug("state", {"data": expensive()})
   end
   ```

5. **Use JSON in production** for machine parsing:
   ```ruby
   log.SetFormat(log.JSONFormat)
   ```
