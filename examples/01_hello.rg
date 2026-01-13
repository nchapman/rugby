# Rugby Hello World
# Demonstrates: bare script mode, puts, string interpolation (spec 2.1, 3.3)

# Top-level statements execute in order (no main function needed)
puts "Hello, Rugby!"

# String interpolation with #{} (compiles to fmt.Sprintf)
name = "World"
puts "Hello, #{name}!"

# Expressions in interpolation
a = 2
b = 3
puts "#{a} + #{b} = #{a + b}"
