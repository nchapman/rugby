# Rugby Error Handling
# Demonstrates: error type, !, rescue, panic
#
# Rugby follows Go's philosophy: errors are values, handled explicitly.

import errors
import os

# Function that can fail - returns (T, Error)
def divide(a: Int, b: Int): (Int, Error)
  return 0, errors.new("division by zero") if b == 0
  return a / b, nil
end

# Propagate errors with ! (like Rust's ?)
def safe_calc(x: Int, y: Int): (Int, Error)
  result = divide(x, y)!
  return result * 2, nil
end

# Function returning just Error
def validate(n: Int): Error
  return errors.new("negative not allowed") if n < 0
  nil
end

def main
  # Explicit error handling (Go-style)
  result, err = divide(10, 2)
  if err == nil
    puts "10 / 2 = #{result}"
  end

  result2, err2 = divide(10, 0)
  if err2 != nil
    puts "Division error caught"
  end

  # Bang operator in main - exits on error
  good = divide(20, 4)!
  puts "20 / 4 = #{good}"

  # rescue with inline default
  safe = divide(10, 0) rescue -1
  puts "10 / 0 with rescue: #{safe}"

  # rescue with block form
  value = divide(5, 0) rescue do
    puts "  Caught error, using fallback"
    0
  end
  puts "Block rescue: #{value}"

  # rescue with error binding
  data = divide(5, 0) rescue => err do
    puts "  Error: #{err}"
    -999
  end
  puts "Error binding: #{data}"

  # Error propagation through call chain
  calc_ok = safe_calc(20, 5) rescue -1
  calc_fail = safe_calc(10, 0) rescue -1
  puts "safe_calc(20, 5): #{calc_ok}"
  puts "safe_calc(10, 0): #{calc_fail}"

  # Validate function (error-only)
  err = validate(10)
  puts "validate(10): #{err == nil ? "ok" : "failed"}"

  err = validate(-5)
  puts "validate(-5): #{err == nil ? "ok" : "failed"}"

  # Error utilities
  _, read_err = os.read_file("/nonexistent/file")
  if error_is?(read_err, os.ErrNotExist)
    puts "File not found (error_is? works)"
  end

  puts "Done!"
end
