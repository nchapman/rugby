# Rugby Hello World
# Demonstrates: bare script mode, string interpolation

# Top-level statements execute in order - no main function needed
puts "Hello, Rugby!"

# String interpolation with #{}
name = "World"
puts "Hello, #{name}!"

# Any expression works inside #{}
a = 2
b = 3
puts "#{a} + #{b} = #{a + b}"

# Method calls in interpolation
greeting = "hello"
puts "Uppercase: #{greeting.upcase}"
