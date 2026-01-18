#@ run-pass
#@ check-output
#
# Test compile-time template.compile with filters

import "rugby/template"

# String filters
const UPCASE = template.compile("{{ name | upcase }}")
const DOWNCASE = template.compile("{{ name | downcase }}")
const CAPITALIZE = template.compile("{{ name | capitalize }}")
const STRIP = template.compile("{{ name | strip }}")

# Array filters
const FIRST = template.compile("{{ items | first }}")
const LAST = template.compile("{{ items | last }}")
const SIZE = template.compile("{{ items | size }}")
const JOIN = template.compile("{{ items | join: \", \" }}")
const REVERSE = template.compile("{{ items | reverse | join: \"-\" }}")

# Filter chains
const CHAIN = template.compile("{{ name | upcase | downcase | capitalize }}")

# Default filter
const DEFAULT = template.compile("{{ missing | default: \"N/A\" }}")

puts UPCASE.MustRender({name: "hello"})
puts DOWNCASE.MustRender({name: "WORLD"})
puts CAPITALIZE.MustRender({name: "john doe"})
puts STRIP.MustRender({name: "  spaces  "})
puts FIRST.MustRender({items: [1, 2, 3]})
puts LAST.MustRender({items: [1, 2, 3]})
puts SIZE.MustRender({items: [1, 2, 3, 4, 5]})
puts JOIN.MustRender({items: ["a", "b", "c"]})
puts REVERSE.MustRender({items: [1, 2, 3]})
puts CHAIN.MustRender({name: "heLLo"})
puts DEFAULT.MustRender({})

#@ expect:
# HELLO
# world
# John doe
# spaces
# 1
# 3
# 5
# a, b, c
# 3-2-1
# Hello
# N/A
