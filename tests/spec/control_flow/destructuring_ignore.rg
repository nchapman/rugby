#@ run-pass
#@ check-output
#
# Test: Section 9.3 - Underscore to ignore values in destructuring

def get_triple: (Int, String, Bool)
  1, "hello", true
end

# Ignore first and third
_, second, _ = get_triple
puts second

# Ignore just the second
first, _, third = get_triple
puts first
puts third

# Can use multiple underscores
def get_five: (Int, Int, Int, Int, Int)
  1, 2, 3, 4, 5
end

a, _, _, _, e = get_five
puts a
puts e

#@ expect:
# hello
# 1
# true
# 1
# 5
