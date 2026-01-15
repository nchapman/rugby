#@ compile-fail
#@ skip: Structs not yet implemented (Section 12.1)
#
# Test: Section 12.1 - Struct definition
# TODO: Implement struct syntax

struct Point
  x : Int
  y : Int
end

p = Point{x: 10, y: 20}
puts p.x
puts p.y
