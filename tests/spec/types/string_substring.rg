#@ run-pass
#@ check-output
# Test String.substring method

s = "hello world"
puts s.substring(0, 5)
puts s.substring(6, 11)

# Single character
puts s.substring(0, 1)

# Empty substring
puts s.substring(5, 5)

# Full string
puts s.substring(0, 11)

#@ expect:
# hello
# world
# h
#
# hello world
