#@ compile-fail
# Bug: Compound assignment on properties fails to parse
# Using += -= etc on property setters causes parse error

class Body
  property vx: Float
  def initialize(@vx: Float)
  end
end

body = Body.new(1.0)

# Expected: body.vx += 0.5 should work
# Actual: "unexpected token +=" parse error
body.vx += 0.5

puts body.vx
#~ ERROR: unexpected token
