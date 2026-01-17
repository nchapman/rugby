#@ run-pass
#@ check-output
#
# Test: Section 11.7 - Designing for abstraction with interfaces

interface Shape
  def area: Float
  def perimeter: Float
end

class Circle implements Shape
  def initialize(@radius: Float)
  end

  def area: Float
    3.14159 * @radius * @radius
  end

  def perimeter: Float
    2.0 * 3.14159 * @radius
  end
end

class Rectangle implements Shape
  def initialize(@width: Float, @height: Float)
  end

  def area: Float
    @width * @height
  end

  def perimeter: Float
    2.0 * (@width + @height)
  end
end

def print_shape(s: Shape)
  puts s.area
  puts s.perimeter
end

print_shape(Circle.new(1.0))
print_shape(Rectangle.new(3.0, 4.0))

#@ expect:
# 3.14159
# 6.28318
# 12
# 14
