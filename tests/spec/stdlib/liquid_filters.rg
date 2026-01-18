#@ run-pass
#@ check-output
#
# Test liquid filters

import "rugby/liquid"

# String filter - upcase
r1 = liquid.render("{{ text | upcase }}", {text: "hello"})!
puts r1

# String filter - downcase
r2 = liquid.render("{{ text | downcase }}", {text: "HELLO"})!
puts r2

# Array filter - first
r3 = liquid.render("{{ items | first }}", {items: [1, 2, 3]})!
puts r3

# Array filter - size
r4 = liquid.render("{{ items | size }}", {items: [1, 2, 3, 4, 5]})!
puts r4

# Default filter
r5 = liquid.render("{{ missing | default: \"N/A\" }}", {})!
puts r5

# Math filter - plus
r6 = liquid.render("{{ num | plus: 5 }}", {num: 10})!
puts r6

#@ expect:
# HELLO
# hello
# 1
# 5
# N/A
# 15
