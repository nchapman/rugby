#@ run-pass
#@ check-output
#
# Test liquid control flow tags

import "rugby/liquid"

# If statement - true
r1 = liquid.render("{% if show %}yes{% endif %}", {show: true})!
puts r1

# If statement - false
r2 = liquid.render("{% if show %}yes{% endif %}", {show: false})!
puts "empty:" + r2 + ":"

# If-else
r3 = liquid.render("{% if x %}on{% else %}off{% endif %}", {x: false})!
puts r3

# For loop
r4 = liquid.render("{% for n in nums %}{{ n }}{% endfor %}", {nums: [1, 2, 3]})!
puts r4

# For loop with range
r5 = liquid.render("{% for i in (1..3) %}{{ i }}{% endfor %}", {})!
puts r5

#@ expect:
# yes
# empty::
# off
# 123
# 123
