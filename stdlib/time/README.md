# rugby/time

Time and duration operations for Rugby programs.

## Import

```ruby
import rugby/time
```

## Quick Start

```ruby
import rugby/time

# Current time
now = time.Now()
puts now.Format("2006-01-02 15:04:05")

# Duration arithmetic
tomorrow = now.Add(time.Hours(24))
yesterday = now.Add(time.Hours(-24))

# Measure elapsed time
start = time.Now()
do_work()
elapsed = time.Since(start)
puts "Took #{elapsed.ToSeconds()} seconds"
```

## Time Functions

### Creating Times

#### time.Now() -> Time

Returns the current local time.

```ruby
now = time.Now()
```

#### time.UTC() -> Time

Returns the current UTC time.

```ruby
utc = time.UTC()
```

#### time.Unix(seconds) -> Time

Creates a time from Unix timestamp (seconds since 1970-01-01 00:00:00 UTC).

```ruby
t = time.Unix(1609459200)  # 2021-01-01 00:00:00 UTC
```

#### time.UnixMilli(ms) -> Time

Creates a time from Unix timestamp in milliseconds.

```ruby
t = time.UnixMilli(1609459200000)
```

#### time.Parse(layout, str) -> (Time, error)

Parses a time string using the given layout.

```ruby
t = time.Parse("2006-01-02", "2024-01-15")!
t = time.Parse("2006-01-02 15:04:05", "2024-01-15 09:30:00")!
```

#### time.ParseRFC3339(str) -> (Time, error)

Parses an RFC3339 formatted time string.

```ruby
t = time.ParseRFC3339("2024-01-15T09:30:00Z")!
```

### Time Methods

#### t.Format(layout) -> String

Formats the time using the given layout.

```ruby
t.Format("2006-01-02")           # "2024-01-15"
t.Format("Mon Jan 2 15:04:05")   # "Mon Jan 15 09:30:00"
t.Format(time.DateTimeLayout)    # "2024-01-15 09:30:00"
```

#### t.RFC3339() -> String

Formats as RFC3339.

```ruby
t.RFC3339()  # "2024-01-15T09:30:00Z"
```

#### t.String() -> String

Returns the time as a string (RFC3339 format).

```ruby
puts t  # Uses String() automatically
```

#### t.Unix() -> Int64

Returns Unix timestamp in seconds.

```ruby
ts = t.Unix()  # 1705309800
```

#### t.UnixMilli() -> Int64

Returns Unix timestamp in milliseconds.

```ruby
ms = t.UnixMilli()  # 1705309800000
```

### Time Components

```ruby
t.Year()     # 2024
t.Month()    # 1 (January)
t.Day()      # 15
t.Hour()     # 9
t.Minute()   # 30
t.Second()   # 0
t.Weekday()  # 1 (Monday, 0=Sunday)
```

### Time Arithmetic

#### t.Add(duration) -> Time

Returns time plus duration.

```ruby
later = t.Add(time.Hours(2))
earlier = t.Add(time.Hours(-2))
```

#### t.Sub(other) -> Duration

Returns duration between two times.

```ruby
diff = t2.Sub(t1)
puts "#{diff.ToHours()} hours apart"
```

### Time Comparisons

```ruby
t1.Before(t2)  # true if t1 is before t2
t1.After(t2)   # true if t1 is after t2
t1.Equal(t2)   # true if times are equal
t1.IsZero()    # true if t1 is zero time
```

### Timezone Conversion

```ruby
t.UTC()    # Convert to UTC
t.Local()  # Convert to local timezone
```

## Duration Functions

### Creating Durations

```ruby
time.Nanoseconds(1000)   # 1000 nanoseconds
time.Microseconds(1000)  # 1000 microseconds
time.Milliseconds(500)   # 500 milliseconds
time.Seconds(30)         # 30 seconds
time.Minutes(5)          # 5 minutes
time.Hours(2)            # 2 hours
time.Days(7)             # 7 days
```

### Duration Methods

#### Conversion

```ruby
d.ToSeconds()       # Float (e.g., 90.5)
d.ToMilliseconds()  # Int64 (e.g., 90500)
d.ToMinutes()       # Float (e.g., 1.5)
d.ToHours()         # Float (e.g., 0.025)
d.String()          # "1m30.5s"
```

#### Arithmetic

```ruby
d1.Add(d2)   # Sum of durations
d1.Sub(d2)   # Difference
d.Mul(3)     # Multiply by integer
d.Abs()      # Absolute value
```

### Measuring Time

#### time.Since(t) -> Duration

Returns duration since time t.

```ruby
start = time.Now()
do_work()
elapsed = time.Since(start)
puts "Took #{elapsed.String()}"
```

#### time.Until(t) -> Duration

Returns duration until time t.

```ruby
deadline = time.Now().Add(time.Minutes(5))
remaining = time.Until(deadline)
puts "#{remaining.ToSeconds()} seconds left"
```

### Sleeping

#### time.Sleep(duration)

Pauses execution for the duration.

```ruby
time.Sleep(time.Seconds(1))
time.Sleep(time.Milliseconds(500))
```

#### time.SleepSeconds(n)

Pauses execution for n seconds (can be fractional).

```ruby
time.SleepSeconds(1.5)
```

## Layout Constants

Go-style layout strings (use actual values from reference time):

```ruby
time.RFC3339Layout   # "2006-01-02T15:04:05Z07:00"
time.RFC822Layout    # "02 Jan 06 15:04 MST"
time.DateLayout      # "2006-01-02"
time.TimeLayout      # "15:04:05"
time.DateTimeLayout  # "2006-01-02 15:04:05"
```

The reference time is: `Mon Jan 2 15:04:05 MST 2006`

## Examples

### Timestamp Logging

```ruby
import rugby/time

def log(message : String)
  ts = time.Now().Format("2006-01-02 15:04:05")
  puts "[#{ts}] #{message}"
end

log("Server started")
```

### Rate Limiting

```ruby
import rugby/time

class RateLimiter
  def initialize(@interval : time.Duration)
    @last = time.Time{}
  end

  def allow? -> Bool
    now = time.Now()
    if @last.IsZero() || time.Since(@last).ToSeconds() >= @interval.ToSeconds()
      @last = now
      return true
    end
    false
  end
end

limiter = RateLimiter.new(time.Seconds(1))
```

### Timeout Pattern

```ruby
import rugby/time

def with_timeout(duration : time.Duration, block : -> any) -> any?
  deadline = time.Now().Add(duration)

  # Check periodically
  while time.Now().Before(deadline)
    result = try_operation()
    if result.ok?
      return result
    end
    time.Sleep(time.Milliseconds(100))
  end

  nil
end
```

### Parse Multiple Formats

```ruby
import rugby/time

def parse_date(s : String) -> time.Time
  formats = [
    "2006-01-02",
    "01/02/2006",
    "Jan 2, 2006",
    time.RFC3339Layout
  ]

  for fmt in formats
    t, err = time.Parse(fmt, s)
    if err == nil
      return t
    end
  end

  panic "Cannot parse date: #{s}"
end
```

### Age Calculation

```ruby
import rugby/time

def age(birthdate : time.Time) -> Int
  now = time.Now()
  years = now.Year() - birthdate.Year()

  # Adjust if birthday hasn't occurred this year
  if now.Month() < birthdate.Month()
    years -= 1
  elsif now.Month() == birthdate.Month() && now.Day() < birthdate.Day()
    years -= 1
  end

  years
end

birthday = time.Parse("2006-01-02", "1990-06-15")!
puts "Age: #{age(birthday)}"
```

### Benchmark

```ruby
import rugby/time

def benchmark(name : String, iterations : Int, block : ->)
  start = time.Now()

  for i in 0...iterations
    block()
  end

  elapsed = time.Since(start)
  per_op = elapsed.ToSeconds() / iterations.to_f

  puts "#{name}: #{elapsed.String()} total, #{per_op * 1000}ms per op"
end

benchmark("sort", 1000) do
  data.sort()
end
```
