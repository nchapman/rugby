#@ compile-fail
# Bug: Class array type mismatch - []*Body vs []Body
# When creating an array of class instances, Go generates []*Body
# but Array<Body> maps to []Body, causing a type mismatch

class Body
  getter x: Float
  def initialize(@x: Float)
  end
end

# This fails because Body.new returns *Body, so the array is []*Body
# but Array<Body> annotation expects []Body
bodies : Array<Body> = [Body.new(1.0), Body.new(2.0)]
i = 0

body : Body = bodies[i]
puts body.x
#~ ERROR: cannot use \[\]\*Body.*as \[\]Body
