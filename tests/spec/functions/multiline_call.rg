#@ run-pass
#@ check-output
# Multi-line function calls work

class Point
  property x: Float
  property y: Float
  property z: Float
  def initialize(@x: Float, @y: Float, @z: Float)
  end
end

# Multi-line constructor call
p = Point.new(
  4.841431,
  -1.160320,
  -0.103622
)

puts p.x
puts p.y
puts p.z

# Multi-line regular function call
def add(a: Int, b: Int, c: Int): Int
  a + b + c
end

result = add(
  1,
  2,
  3
)

puts result

#@ expect:
# 4.841431
# -1.16032
# -0.103622
# 6
