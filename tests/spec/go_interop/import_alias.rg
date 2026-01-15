#@ run-pass
#@ check-output
#
# Test: Section 18.1 - Import with alias

import "strings" as str

s = "hello"
puts str.to_upper(s)
puts str.repeat(s, 3)

#@ expect:
# HELLO
# hellohellohello
