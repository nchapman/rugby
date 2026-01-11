# Error Handling in Rugby
# Demonstrates: error type, ! operator, rescue, and raise

import errors

def might_fail(n : Int) -> (Int, error)
  if n < 0
    return 0, errors.New("negative number")
  end
  return n * 2, nil
end

def main
  # Bang operator - exits with runtime.Fatal on error
  result = might_fail(5)!
  puts("Result: #{result}")

  # Rescue with default value
  safe = might_fail(-1) rescue 0
  puts("Safe result: #{safe}")

  # Raise for panics (commented out)
  # raise "something went wrong!"

  puts("Done!")
end
