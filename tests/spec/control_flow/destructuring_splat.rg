#@ run-pass
#@ check-output
#
# Test: Section 9.3 - Splat patterns in destructuring

items = [1, 2, 3, 4, 5]

# First element, rest in array
first1, *rest1 = items
puts first1
puts rest1.length

# All but last in array, last element
*head2, last2 = items
puts head2.length
puts last2

# First, middle (rest), last
first3, *middle3, last3 = items
puts first3
puts middle3.length
puts last3

#@ expect:
# 1
# 4
# 4
# 5
# 1
# 3
# 5
