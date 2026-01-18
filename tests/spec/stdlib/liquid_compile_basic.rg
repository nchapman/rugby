#@ run-pass
#@ check-output
#
# Test compile-time liquid templates with liquid.compile

import "rugby/liquid"

# Simple text
const HELLO = liquid.compile("Hello!")

# Variable interpolation
const GREETING = liquid.compile("Hello, {{ name }}!")

# Multiple variables
const MESSAGE = liquid.compile("{{ greeting }}, {{ name }}!")

# With filter
const UPPER = liquid.compile("Hello, {{ name | upcase }}!")

puts HELLO.MustRender({})
puts GREETING.MustRender({name: "World"})
puts MESSAGE.MustRender({greeting: "Hi", name: "Alice"})
puts UPPER.MustRender({name: "bob"})

#@ expect:
# Hello!
# Hello, World!
# Hi, Alice!
# Hello, BOB!
