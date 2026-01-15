#@ run-pass
#@ check-output
#
# Test regex module from rugby/regex

import rugby/regex

# Match - check if pattern exists
puts regex.Match("\\d+", "abc123")
puts regex.Match("\\d+", "abc")

# Find - get first match
match, ok = regex.Find("\\d+", "abc123def")
puts ok
puts match

# FindAll - get all matches
matches = regex.FindAll("\\d+", "a1b2c3")
matches.each -> (m) { puts m }

# Replace - replace all matches
result = regex.Replace("hello world", "\\w+", "X")
puts result

#@ expect:
# true
# false
# true
# 123
# 1
# 2
# 3
# X X
