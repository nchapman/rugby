#@ run-pass
#@ check-output
#
# Test basic liquid template rendering

import "rugby/liquid"

# Simple variable output
result1 = liquid.render("Hello, {{ name }}!", {name: "World"})!
puts result1

# Integer in template
result2 = liquid.render("Count: {{ x }}", {x: 42})!
puts result2

# Dot access
result3 = liquid.render("{{ user.name }}", {user: {name: "Alice"}})!
puts result3

# Parse once, render multiple times
tmpl = liquid.parse("Hi, {{ name }}!")!
result4 = tmpl.render({name: "Bob"})!
puts result4

#@ expect:
# Hello, World!
# Count: 42
# Alice
# Hi, Bob!
