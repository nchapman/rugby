#@ run-pass
#@ check-output
#
# Test: Section 10.10 - Calling convention (optional parens)

def greet(name: String)
  puts "Hello, #{name}!"
end

def get_value: Int
  42
end

# With parentheses
greet("World")

# Without parentheses (command syntax)
greet "Ruby"

# No-arg call without parens
x = get_value
puts x

# puts without parens
puts "done"

#@ expect:
# Hello, World!
# Hello, Ruby!
# 42
# done
