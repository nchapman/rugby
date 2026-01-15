#@ run-pass
#@ check-output
#
# Test: Section 18.1 - Go imports

import "strings"
import "strconv"

# Call Go functions with snake_case (mapped to CamelCase)
s = "hello world"
puts strings.to_upper(s)
puts strings.contains(s, "world")

# strconv
n, _ = strconv.atoi("42")
puts n

#@ expect:
# HELLO WORLD
# true
# 42
