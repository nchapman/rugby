#@ run-pass
#@ check-output
#
# Test existing primitive type annotations

# Integer types
def test_int(x: Int): Int
  x * 2
end

def test_int64(x: Int64): Int64
  x + 1
end

# Float type
def test_float(x: Float): Float
  x * 2.0
end

# String type
def test_string(s: String): String
  s + "!"
end

# Bool type
def test_bool(b: Bool): Bool
  not b
end

# Test calls
a: Int = 10
puts test_int(a)

b: Int64 = 100
puts test_int64(b)

c: Float = 1.5
puts test_float(c)

puts test_string("hello")

puts test_bool(true)

#@ expect:
# 20
# 101
# 3
# hello!
# false
