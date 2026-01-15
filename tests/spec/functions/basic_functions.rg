#@ run-pass
#@ check-output
#
# Test: Section 10.1 - Basic functions

# Implicit return (last expression)
def add(a : Int, b : Int) -> Int
  a + b
end

puts add(2, 3)

# Explicit return
def find_first_positive(arr : Array<Int>) -> Int
  for n in arr
    return n if n > 0
  end
  0
end

puts find_first_positive([-1, -2, 5, 10])
puts find_first_positive([-1, -2])

# No return type (inferred)
def greet(name : String)
  puts "Hello, #{name}!"
end

greet("World")

#@ expect:
# 5
# 5
# 0
# Hello, World!
