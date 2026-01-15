#@ run-pass
#@ check-output
#
# Test: Section 18.2 - Name mapping (snake_case to CamelCase)

import "strings"

# snake_case in Rugby -> CamelCase in Go
# has_prefix -> HasPrefix
# to_lower -> ToLower
# replace_all -> ReplaceAll

s = "Hello World"

puts strings.has_prefix(s, "Hello")
puts strings.to_lower(s)
puts strings.replace_all(s, "o", "0")

#@ expect:
# true
# hello world
# Hell0 W0rld
