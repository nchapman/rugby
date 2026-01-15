#@ run-pass
#@ check-output
#
# Test: Section 5.1 - Float primitive types

# Float (float64)
x : Float = 3.14159
puts x

# Float operations
puts 1.5 + 2.5
puts 5.0 - 2.0
puts 2.5 * 4.0
puts 10.0 / 4.0

# Float predicates
f = 0.0
puts f.zero?

f2 = 3.14
puts f2.positive?

#@ expect:
# 3.14159
# 4
# 3
# 10
# 2.5
# true
# true
