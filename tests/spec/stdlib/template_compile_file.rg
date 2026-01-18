#@ run-pass
#@ check-output
#
# Test compile-time template.compile with compile_file

import "rugby/template"

# Load template from file
const TEMPLATE = template.compile_file("template/test_template.tmpl")

puts TEMPLATE.MustRender({name: "World", items: [1, 2, 3]})

#@ expect:
# Hello, World!
# Your items: 1, 2, 3
