#@ compile-fail
#@ skip: Structs not yet implemented (Section 12.4)
#
# Test: Section 12.4 - Struct field assignment compile error
# TODO: Implement struct immutability enforcement
#
# This test should fail to compile because struct fields cannot be assigned

struct Point
  x : Int
  y : Int
end

p = Point{x: 10, y: 20}
p.x = 5  # ERROR: cannot modify struct field

#~ ERROR: cannot modify struct field
