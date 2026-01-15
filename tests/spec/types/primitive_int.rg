#@ run-pass
#@ check-output
#
# Test: Section 5.1 - Integer primitive types

# Int (platform-dependent)
x : Int = 42
puts x

# Int64
big : Int64 = 9223372036854775807
puts big

# Negative numbers
neg : Int = -100
puts neg

# Integer operations
puts 10 + 5
puts 10 - 3
puts 4 * 5
puts 20 / 4
puts 17 % 5

#@ expect:
# 42
# 9223372036854775807
# -100
# 15
# 7
# 20
# 5
# 2
