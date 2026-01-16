#@ run-pass
#@ check-output
#
# Test: Section 6.4 - T.zero type-level method
# Returns the zero value for a Numeric type parameter

def zero_like<T : Numeric>(x : T) -> T
  T.zero
end

# Works with integers
puts zero_like(42)
puts zero_like(-100)

# Works with floats
puts zero_like(3.14)
puts zero_like(-2.5)

#@ expect:
# 0
# 0
# 0
# 0
