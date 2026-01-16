#@ run-pass
#@ check-output
#
# Test: Section 12.3 - Struct methods

import "math"

struct Point
  x : Int
  y : Int

  def distance_to(other : Point) -> Float
    dx = (other.x - @x).to_f
    dy = (other.y - @y).to_f
    math.Sqrt(dx * dx + dy * dy)
  end

  def translate(dx : Int, dy : Int) -> Point
    Point{x: @x + dx, y: @y + dy}
  end
end

p1 = Point{x: 0, y: 0}
p2 = Point{x: 3, y: 4}
puts p1.distance_to(p2)

p3 = p1.translate(5, 10)
puts p3

#@ expect:
# 5
# Point{x: 5, y: 10}
