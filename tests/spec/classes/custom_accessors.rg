#@ run-pass
#@ check-output
#
# Test: Section 11.4 - Custom accessors

class Temperature
  def initialize(@celsius : Float)
  end

  getter celsius : Float

  def fahrenheit -> Float
    @celsius * 9.0 / 5.0 + 32.0
  end

  def fahrenheit=(f : Float)
    @celsius = (f - 32.0) * 5.0 / 9.0
  end
end

t = Temperature.new(0.0)
puts t.celsius
puts t.fahrenheit

t.fahrenheit = 212.0
puts t.celsius

#@ expect:
# 0
# 32
# 100
