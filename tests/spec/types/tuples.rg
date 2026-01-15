#@ run-pass
#@ check-output
#
# Test tuple literals and multi-value returns (Section 5.4)

# Explicit return with multiple values
def explicit_pair -> (Int, String)
  return 42, "answer"
end

# Implicit return with tuple literal
def implicit_pair -> (Int, String)
  100, "hundred"
end

# Multi-value return with more elements
def get_bounds -> (Int, Int, Int, Int)
  0, 0, 800, 600
end

# Conditional return with tuples
def maybe_value(n : Int) -> (Int, Bool)
  return 0, false if n < 0
  n * 2, true
end

def main
  # Test explicit return
  a, b = explicit_pair
  puts a
  puts b

  # Test implicit return
  c, d = implicit_pair
  puts c
  puts d

  # Test 4-tuple
  x, y, w, h = get_bounds
  puts x
  puts y
  puts w
  puts h

  # Test conditional tuple return
  n1, ok1 = maybe_value(50)
  puts n1
  puts ok1

  n2, ok2 = maybe_value(-5)
  puts n2
  puts ok2
end

#@ expect:
# 42
# answer
# 100
# hundred
# 0
# 0
# 800
# 600
# 100
# true
# 0
# false
