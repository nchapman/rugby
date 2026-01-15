#@ run-pass
#@ check-output
#@ skip: Destructuring not yet implemented (Section 9.3)
#
# Test: Section 9.3 - Splat patterns in destructuring
# TODO: Implement splat destructuring

items = [1, 2, 3, 4, 5]

# First element, rest in array
first, *rest = items
puts first
puts rest.length

# All but last in array, last element
*head, last = items
puts head.length
puts last

# First, middle (rest), last
first, *middle, last = items
puts first
puts middle.length
puts last

#@ expect:
# 1
# 4
# 4
# 5
# 1
# 3
# 5
