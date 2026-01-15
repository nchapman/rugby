#@ run-pass
#@ check-output
#@ skip: Structs not yet implemented (Section 12.4)
#
# Test: Section 12.4 - Correct pattern for "modifying" structs
# TODO: Implement structs
#
# Since structs are immutable, methods return new structs instead of mutating

struct Point
  x : Int
  y : Int

  def moved(dx : Int, dy : Int) -> Point
    Point{x: @x + dx, y: @y + dy}
  end

  def scaled(factor : Int) -> Point
    Point{x: @x * factor, y: @y * factor}
  end
end

p = Point{x: 10, y: 20}
q = p.moved(5, 5)

puts p.x  # Original unchanged
puts p.y
puts q.x  # New point with moved values
puts q.y

r = p.scaled(2)
puts r.x
puts r.y

#@ expect:
# 10
# 20
# 15
# 25
# 20
# 40
