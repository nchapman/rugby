#@ run-pass
#@ check-output
#
# Test: Section 6.3 - Generic interfaces

interface Container<T>
  def get -> T
  def set(value : T)
end

class Box<T>
  def initialize(@value : T)
  end

  def get -> T
    @value
  end

  def set(value : T)
    @value = value
  end
end

# Box implements Container<Int>
box = Box.new(42)
puts box.get

box.set(100)
puts box.get

#@ expect:
# 42
# 100
