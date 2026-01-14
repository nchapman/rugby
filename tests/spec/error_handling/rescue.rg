#@ run-pass
#@ check-output
#
# Test rescue expressions for error handling

import errors

# Function that returns (value, error)
def parse(input : String) -> (Int, error)
  if input == "bad"
    return 0, errors.New("invalid input")
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
