#@ run-pass
#@ check-output
#
# Test: Section 4.4 - Single quote strings (no interpolation)

# Single quotes are literal - no interpolation
name = "World"
puts 'Hello #{name}'

# Only \' and \\ are escape sequences in single quotes
puts 'it\'s fine'
puts 'path\\to\\file'

# Double quotes for comparison - interpolation works
puts "Hello #{name}"

#@ expect:
# Hello #{name}
# it's fine
# path\to\file
# Hello World
