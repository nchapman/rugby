#@ compile-fail
#
# Test: Section 12.4 - Struct immutability

struct Point
  x : Int
  y : Int
end

def main
  p = Point{x: 10, y: 20}
  p.x = 5  # Should be compile error: cannot modify struct field
end

#~ ERROR: cannot modify struct field
