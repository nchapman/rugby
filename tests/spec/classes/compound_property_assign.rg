#@ run-pass
#@ check-output
# Compound assignment on properties works
# Using += -= etc on property setters generates setter calls

class Body
  property vx: Float
  property vy: Float
  def initialize(@vx: Float, @vy: Float)
  end
end

body = Body.new(1.0, 2.0)

# Compound assignment on properties
body.vx += 0.5
body.vy -= 0.3

puts body.vx
puts body.vy

#@ expect:
# 1.5
# 1.7
