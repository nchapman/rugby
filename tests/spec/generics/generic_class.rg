#@ run-pass
#@ check-output
#
# Test: Section 6.2 - Generic classes

class Box<T>
  def initialize(@value: T)
  end

  def get: T
    @value
  end
end

# For now, use type inference (no explicit type param on constructor)
box = Box.new(42)
puts box.get

#@ expect:
# 42
