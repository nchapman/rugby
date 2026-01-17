#@ run-pass
#@ check-output
# Class array with explicit type annotation
# Array<Body> should generate []*Body, not []Body

class Body
  getter x: Float
  def initialize(@x: Float)
  end
end

# Explicit type annotation - should work with class pointer types
bodies : Array<Body> = [Body.new(1.0), Body.new(2.0), Body.new(3.0)]
i = 0
while i < 3
  body = bodies[i]
  puts body.x
  i += 1
end

# Without type annotation also works
bodies2 = [Body.new(4.0), Body.new(5.0)]
puts bodies2[0].x
puts bodies2[1].x

#@ expect:
# 1
# 2
# 3
# 4
# 5
