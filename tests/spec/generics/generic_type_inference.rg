#@ run-pass
#@ check-output
#
# Test: Section 6.5 - Type inference for generics

# Generic class for testing explicit type params
class Box<T>
  def initialize(@value: T)
  end

  def get: T
    @value
  end
end

# Type parameters inferred from arguments
def identity<T>(x: T): T
  x
end

# Inferred as Int
result1 = identity(42)
puts result1

# Inferred as String
result2 = identity("hello")
puts result2

# Type inferred from argument
box = Box.new(42)
puts box.get

# Empty array needs explicit type
result: Array<String> = []
puts result.length

#@ expect:
# 42
# hello
# 42
# 0
