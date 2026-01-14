#@ run-pass
#@ check-output
#
# Test: Go standard library interop

import "strings"

s = "hello world"

# Go string methods (snake_case maps to CamelCase)
puts strings.to_upper(s)
puts strings.contains(s, "world")

# Split and join
parts = strings.split("a,b,c", ",")
puts strings.join(parts, "-")

#@ expect:
# HELLO WORLD
# true
# a-b-c
