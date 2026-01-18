#@ run-pass
#@ check-output
#
# Test compile-time liquid templates with compile_file

import "rugby/liquid"

# Load template from file
const TEMPLATE = liquid.compile_file("liquid/test_template.liquid")

puts TEMPLATE.MustRender({name: "World", items: [1, 2, 3]})

#@ expect:
# Hello, World!
# Your items: 1, 2, 3
