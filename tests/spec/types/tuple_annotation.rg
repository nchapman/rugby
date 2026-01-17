#@ run-pass
#@ check-output

# Test tuple type annotations and destructuring

def get_pair: (String, Int)
  ("hello", 42)
end

def main
  # Tuple return with immediate destructuring (preferred pattern)
  a, b = get_pair
  puts a
  puts b

  # Tuple literal with type annotation
  t: (Int, Int) = (10, 20)
  x, y = t
  puts x
  puts y

  # Tuple reassignment
  t = (30, 40)
  p, q = t
  puts p
  puts q
end

#@ expect:
# hello
# 42
# 10
# 20
# 30
# 40
