#@ run-pass
#@ check-output
#
# Test: Integer literals and basic operations

# Basic integers
puts 42
puts -17
puts 0

# Arithmetic
puts 2 + 3
puts 10 - 4
puts 3 * 7
puts 20 / 4

# Precedence
puts 2 + 3 * 4

# Parenthesized expressions (as separate step)
x = (2 + 3) * 4
puts x

#@ expect:
# 42
# -17
# 0
# 5
# 6
# 21
# 5
# 14
# 20
