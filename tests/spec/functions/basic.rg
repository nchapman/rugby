#@ run-pass
#@ check-output
#
# Test: Basic function definitions and calls

def greet(name : String) -> String
  "Hello, #{name}!"
end

def add(a : Int, b : Int) -> Int
  a + b
end

def no_return
  puts "side effect"
end

puts greet("World")
puts add(2, 3)
no_return

#@ expect:
# Hello, World!
# 5
# side effect
