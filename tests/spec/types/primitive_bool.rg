#@ run-pass
#@ check-output
#
# Test: Section 5.1 - Bool primitive type

t: Bool = true
f: Bool = false

puts t
puts f

# Boolean operators
puts true && true
puts true && false
puts false || true
puts !false

# Comparison produces Bool
puts 5 > 3
puts 5 < 3
puts 5 == 5
puts 5 != 3

#@ expect:
# true
# false
# true
# false
# true
# true
# true
# false
# true
# true
