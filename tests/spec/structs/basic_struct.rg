#@ run-pass
#@ check-output
#
# Test: Section 12.1 - Struct definition

struct Point
  x : Int
  y : Int
end

p = Point{x: 10, y: 20}
puts p.x
puts p.y

#@ expect:
# 10
# 20
