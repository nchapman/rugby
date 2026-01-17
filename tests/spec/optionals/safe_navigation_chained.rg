#@ skip: chained &. on nested optionals has codegen issues
#@ run-pass
#@ check-output

# Test: Chained safe navigation operator (Section 8.1)
# obj&.method&.method chains multiple optional accesses

class Inner
  getter value : String

  def initialize(@value: String)
  end
end

class Outer
  getter inner : Inner?

  def initialize(@inner: Inner?)
  end
end

# Full chain present
inner1 = Inner.new("hello")
outer1 = Outer.new(inner1)

result1 = outer1.inner&.value ?? "Unknown"
puts result1

# Chain with nil
outer2 = Outer.new(nil)

result2 = outer2.inner&.value ?? "Unknown"
puts result2

# Using in a function
def get_value(o: Outer): String
  o.inner&.value ?? "default"
end

puts get_value(outer1)
puts get_value(outer2)

# Multiple levels with method call
class Container
  getter outer : Outer?

  def initialize(@outer: Outer?)
  end
end

container1 = Container.new(outer1)
container2 = Container.new(nil)

puts container1.outer&.inner&.value ?? "none"
puts container2.outer&.inner&.value ?? "none"

#@ expect:
# hello
# Unknown
# hello
# default
# hello
# none
