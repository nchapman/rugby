#@ run-pass
#@ check-output
#
# Test compile-time template.compile with template.compile

import "rugby/template"

# Simple text
const HELLO = template.compile("Hello!")

# Variable interpolation
const GREETING = template.compile("Hello, {{ name }}!")

# Multiple variables
const MESSAGE = template.compile("{{ greeting }}, {{ name }}!")

# With filter
const UPPER = template.compile("Hello, {{ name | upcase }}!")

puts HELLO.MustRender({})
puts GREETING.MustRender({name: "World"})
puts MESSAGE.MustRender({greeting: "Hi", name: "Alice"})
puts UPPER.MustRender({name: "bob"})

#@ expect:
# Hello!
# Hello, World!
# Hi, Alice!
# Hello, BOB!
