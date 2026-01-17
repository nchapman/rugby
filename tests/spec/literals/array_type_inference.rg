#@ run-pass
#@ check-output
# Array type inference from default value in Array.new(size, default)

# Int inferred from 0
nums = Array.new(3, 0)
nums[0] = 42
puts nums[0]

# Float inferred from 0.0
floats = Array.new(2, 0.0)
floats[0] = 3.14
puts floats[0]

# Bool inferred from false
flags = Array.new(2, false)
flags[0] = true
puts flags[0]

# String inferred from ""
names = Array.new(2, "")
names[0] = "hello"
puts names[0]

# Class instance as default
class Point
  property x: Int
  property y: Int

  def initialize(@x: Int, @y: Int)
  end
end

points = Array.new(2, Point.new(0, 0))
points[0] = Point.new(3, 4)
puts points[0].x

# Expression as default (function call)
def default_value: Int
  100
end

scores = Array.new(3, default_value)
puts scores[1]

# Array.new(size) without default -> Array<any>
mixed = Array.new(3)
mixed[0] = 1
mixed[1] = "two"
mixed[2] = true
puts mixed.length

#@ expect:
# 42
# 3.14
# true
# hello
# 3
# 100
# 3
