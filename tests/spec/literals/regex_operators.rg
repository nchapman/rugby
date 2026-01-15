#@ run-pass
#@ check-output
#
# Test: Section 4.7 - Regular expression operators

import "rugby/regex"

text = "hello world"

# =~ returns true if match
if regex.match("hello", text)
  puts "matched hello"
end

# Check for digits
if regex.match("\\d+", "abc123")
  puts "has digits"
end

# No match
if !regex.match("xyz", text)
  puts "no xyz"
end

#@ expect:
# matched hello
# has digits
# no xyz
