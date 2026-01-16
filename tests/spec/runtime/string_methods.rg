#@ run-pass
#@ check-output
#
# Test: Section 19.3 - String methods

s = "  Hello World  "

# Query
puts s.length
puts s.empty?
puts "".empty?
puts s.include?("World")
puts s.start_with?("  Hello")
puts s.end_with?("  ")

# Transformation
puts s.upcase
puts s.downcase
puts s.strip

# Split
parts = "a,b,c".split(",")
puts parts.length
puts parts[0]
puts parts[1]

#@ expect:
# 15
# false
# true
# true
# true
# true
#   HELLO WORLD  
#   hello world  
# Hello World
# 3
# a
# b
