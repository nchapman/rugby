#@ run-pass
#@ check-output
#
# Test: Section 5.4 - Tuple types

# Function returning tuple
def get_pair -> (Int, String)
  42, "hello"
end

# Destructuring tuple return
n, s = get_pair
puts n
puts s

# Ignore second element with _
first, _ = get_pair
puts first

# Multiple values
def get_bounds -> (Int, Int, Int, Int)
  0, 0, 100, 200
end

x, y, w, h = get_bounds
puts x
puts y
puts w
puts h

#@ expect:
# 42
# hello
# 42
# 0
# 0
# 100
# 200
