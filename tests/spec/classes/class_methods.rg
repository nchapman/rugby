#@ run-pass
#@ check-output
#
# Test class methods (def self.method)
#
# Class methods are methods that belong to the class itself rather than
# instances. They're commonly used for:
# - Factory methods (e.g., User.create, Point.origin)
# - Utility functions (e.g., Math.sqrt, JSON.parse)
# - Class-level state access

class Point
  getter x : Int
  getter y : Int

  def initialize(@x : Int, @y : Int)
  end

  # Class method to create origin point
  def self.origin -> Point
    Point.new(0, 0)
  end

  # Class method with parameters
  def self.from_coords(x : Int, y : Int) -> Point
    Point.new(x, y)
  end
end

# Should be able to call class methods on the class itself
origin = Point.origin
puts origin.x
puts origin.y

point = Point.from_coords(3, 4)
puts point.x
puts point.y

#@ expect:
# 0
# 0
# 3
# 4
