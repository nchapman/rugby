#@ compile-fail
#@ skip: Structs not yet implemented (Section 12.3)
#
# Test: Section 12.3 - Struct methods
# TODO: Implement struct methods

struct Point
  x : Int
  y : Int

  def distance_to(other : Point) -> Float
    dx = (other.x - @x).to_f
    dy = (other.y - @y).to_f
    Math.sqrt(dx * dx + dy * dy)
  end

  def translate(dx : Int, dy : Int) -> Point
    Point{x: @x + dx, y: @y + dy}
  end
end

p1 = Point{x: 0, y: 0}
p2 = Point{x: 3, y: 4}
puts p1.distance_to(p2)
