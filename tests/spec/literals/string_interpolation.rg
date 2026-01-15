#@ run-pass
#@ check-output
#
# Test: Section 4.4 - String interpolation with #{}

name = "Rugby"
count = 42
price = 19.99

# Basic interpolation
puts "Hello, #{name}!"

# Expression interpolation
puts "count: #{count}"
puts "sum: #{1 + 2 + 3}"
puts "double: #{count * 2}"

# Multiple interpolations
puts "#{name} costs $#{price}"

# Nested method calls
puts "length: #{name.length}"

#@ expect:
# Hello, Rugby!
# count: 42
# sum: 6
# double: 84
# Rugby costs $19.99
# length: 5
