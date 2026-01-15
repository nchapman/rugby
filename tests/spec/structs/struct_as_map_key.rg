#@ run-pass
#@ check-output
#@ skip: Structs not yet implemented (Section 12.2)
#
# Test: Section 12.2 - Structs as map keys (requires hash code)
# TODO: Implement struct hash code generation

struct Point
  x : Int
  y : Int
end

# Structs can be used as map keys because they have hash codes
points = Map<Point, String>{}

p1 = Point{x: 0, y: 0}
p2 = Point{x: 1, y: 1}
p3 = Point{x: 0, y: 0}  # Same as p1

points[p1] = "origin"
points[p2] = "diagonal"

# p3 equals p1, so it should retrieve the same value
puts points[p3]

# Update via equal key
points[p3] = "updated origin"
puts points[p1]

puts points.length

#@ expect:
# origin
# updated origin
# 2
