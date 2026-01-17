#@ run-pass
#@ check-output
#
# Test: Section 5.1 - String primitive type (immutable UTF-8)

s: String = "hello"
puts s

# String methods
puts s.length
puts s.upcase
puts s.downcase

# String concatenation
puts "hello" + " " + "world"

# String predicates
puts "".empty?
puts "hello".empty?
puts "hello".include?("ell")

#@ expect:
# hello
# 5
# HELLO
# hello
# hello world
# true
# false
# true
