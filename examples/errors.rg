# Error Handling in Rugby
# Demonstrates: error type, ! operator, rescue, and panic

def might_fail(n : Int) -> (Int, error)
  if n < 0
    return 0, error("negative number")
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

  # Panic for unrecoverable errors (commented out)
  # panic "something went wrong!"

  puts("Done!")
end
