#@ run-pass
#@ check-output
#
# Test: String literals and interpolation

# Basic strings
puts "hello"
puts 'world'

# Interpolation
name = "Rugby"
puts "Hello, #{name}!"

# Escapes
puts "line1\nline2"
puts "tab\there"

#@ expect:
# hello
# world
# Hello, Rugby!
# line1
# line2
# tab	here
