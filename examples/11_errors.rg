# Rugby Error Handling
# Demonstrates: error type, !, rescue, panic, error utilities (spec 15)

import errors
import os

# Function that can fail (spec 15.1)
def divide(a : Int, b : Int) -> (Int, error)
  if b == 0
    return 0, errors.New("division by zero")
  end
  return a / b, nil
end

# Function that propagates errors with ! (spec 15.2)
def safe_calc(x : Int, y : Int) -> (Int, error)
  result = divide(x, y)!
  return result * 2, nil
end

# Function returning just error (spec 15.1)
def validate(n : Int) -> error
  if n < 0
    return errors.New("negative not allowed")
  end
  nil
end

# panic for unrecoverable errors (spec 15.6)
def must_be_positive(n : Int)
  panic "value must be positive, got #{n}" if n <= 0
end

def main
  # Explicit error handling (spec 15.5)
  result, err = divide(10, 2)
  if err == nil
    puts "10 / 2 = #{result}"
  end

  result2, err2 = divide(10, 0)
  if err2 != nil
    puts "Error: division by zero"
  end

  # Bang operator in main - exits on error (spec 15.4)
  good = divide(20, 4)!
  puts "20 / 4 = #{good}"

  # rescue with default value (spec 15.3)
  safe = divide(10, 0) rescue -1
  puts "10 / 0 with rescue: #{safe}"

  # rescue block form
  value = divide(5, 0) rescue do
    puts "  Caught error, using fallback"
    0
  end
  puts "Value with block rescue: #{value}"

  # rescue with error binding (spec 15.3)
  data = divide(5, 0) rescue => err do
    puts "  Error occurred: #{err}"
    -999
  end
  puts "With error binding: #{data}"

  # Error propagation
  calc_result = safe_calc(20, 5) rescue -1
  puts "safe_calc(20, 5): #{calc_result}"

  calc_fail = safe_calc(10, 0) rescue -1
  puts "safe_calc(10, 0): #{calc_fail}"

  # Validate function (error-only return)
  err = validate(10)
  if err == nil
    puts "10 is valid"
  end

  err = validate(-5)
  if err != nil
    puts "Validation failed for -5"
  end

  # Error utilities (spec 15.7)
  _, read_err = os.read_file("/nonexistent/file")
  if error_is?(read_err, os.ErrNotExist)
    puts "File not found (error_is? works)"
  end

  # must_be_positive(5) works fine
  # must_be_positive(-1) would panic

  puts "Done!"
end
