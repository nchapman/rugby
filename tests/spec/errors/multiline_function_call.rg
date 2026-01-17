#@ compile-fail
# Bug: Multi-line function calls parse errors
# Function calls split across multiple lines fail to parse

class Point
  def initialize(@x: Float, @y: Float, @z: Float)
  end
end

# Expected behavior: should work
# p = Point.new(
#   4.841431,
#   -1.160320,
#   -0.103622
# )

# This fails:
p = Point.new(
  4.841431,
  -1.160320,
  -0.103622
)
#~ ERROR: expected
