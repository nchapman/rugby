#@ run-pass
#@ check-output
#
# Test compile-time liquid templates with filters

import "rugby/liquid"

# String filters
const UPCASE = liquid.compile("{{ name | upcase }}")
const DOWNCASE = liquid.compile("{{ name | downcase }}")
const CAPITALIZE = liquid.compile("{{ name | capitalize }}")
const STRIP = liquid.compile("{{ name | strip }}")

# Array filters
const FIRST = liquid.compile("{{ items | first }}")
const LAST = liquid.compile("{{ items | last }}")
const SIZE = liquid.compile("{{ items | size }}")
const JOIN = liquid.compile("{{ items | join: \", \" }}")
const REVERSE = liquid.compile("{{ items | reverse | join: \"-\" }}")

# Filter chains
const CHAIN = liquid.compile("{{ name | upcase | downcase | capitalize }}")

# Default filter
const DEFAULT = liquid.compile("{{ missing | default: \"N/A\" }}")

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
