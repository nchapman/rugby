#@ run-pass
#@ check-output
#
# Test float literals and operations

# Basic float literals
a = 3.14
b = 2.5
puts a
puts b

# Float arithmetic
puts a + b
puts a * 2.0

# Integer to float conversion
n = 5
puts n.to_f

# Float predicates
puts 0.0.zero?
puts 3.14.positive?
neg = -2.5
puts neg.negative?

# Rounding operations
x = 3.7
puts x.floor
puts x.ceil
puts x.round

# Very small numbers
tiny = 0.0001
puts tiny

# Float comparison
puts 1.0 == 1.0

#@ expect:
# 3.14
# 2.5
# 5.640000000000001
# 6.28
# 5
# true
# true
# true
# 3
# 4
# 4
# 0.0001
# true
