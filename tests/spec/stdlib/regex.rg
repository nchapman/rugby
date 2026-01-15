#@ run-pass
#@ check-output
#
# Test regex module from rugby/regex

import "rugby/regex"

# match - check if pattern exists
puts regex.match("\\d+", "abc123")
puts regex.match("\\d+", "abc")

# find - get first match
match, ok = regex.find("\\d+", "abc123def")
puts ok
puts match

# find_all - get all matches
matches = regex.find_all("\\d+", "a1b2c3")
matches.each -> (m) { puts m }

# replace - replace all matches
result = regex.replace("hello world", "\\w+", "X")
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
