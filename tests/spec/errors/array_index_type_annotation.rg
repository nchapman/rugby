#@ compile-fail
# Bug: Type assertion needed for array index even with explicit annotation
# Array<T>[i] returns 'any', requiring type assertion even when
# the target variable has explicit type annotation

class Body
  getter x: Float
  def initialize(@x: Float)
  end
end

bodies : Array<Body> = [Body.new(1.0), Body.new(2.0)]
i = 0

# Expected: bodies[i] should be Body since Array<Body>
# Actual: bodies[i] is 'any', explicit annotation doesn't help
body : Body = bodies[i]

puts body.x
#~ ERROR: need type assertion
