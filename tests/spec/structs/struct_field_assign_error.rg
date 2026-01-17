#@ compile-fail
#
# Test: Section 12.4 - Struct field assignment compile error
# This test should fail to compile because struct fields cannot be assigned

struct Point
  x: Int
  y: Int
end

def main
  p = Point{x: 10, y: 20}
  p.x = 5  # ERROR: cannot modify struct field
end

#~ ERROR: cannot modify struct field
