#@ run-pass
#@ check-output
#
# Test: Nil coalescing operator (??)

def maybe_string(return_nil: Bool): String?
  return nil if return_nil
  "value"
end

# With value
result1 = maybe_string(false) ?? "default"
puts result1

# With nil
result2 = maybe_string(true) ?? "default"
puts result2

#@ expect:
# value
# default
