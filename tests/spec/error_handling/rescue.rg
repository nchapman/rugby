#@ run-pass
#@ check-output
#
# Test: Section 16.5 - rescue expressions for error handling

import "errors"

# Function that returns (value, Error)
def parse(input: String): (Int, Error)
  if input == "bad"
    return 0, errors.new("invalid input")
  end
  return 42, nil
end

# Inline rescue with default value
good = parse("ok") rescue -1
puts good

bad = parse("bad") rescue -1
puts bad

# Rescue with block
result = parse("bad") rescue do
  puts "handling error"
  999
end
puts result

#@ expect:
# 42
# -1
# handling error
# 999
