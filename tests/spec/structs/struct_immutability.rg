#@ compile-fail
#@ skip: Structs not yet implemented (Section 12.4)
#
# Test: Section 12.4 - Struct immutability
# TODO: Implement struct immutability

struct Point
  x : Int
  y : Int
end

p = Point{x: 10, y: 20}
p.x = 5  # Should be compile error: cannot modify struct field
