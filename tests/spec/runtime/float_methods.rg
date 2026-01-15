#@ run-pass
#@ check-output
#
# Test: Section 19.5 - Float methods

f = 3.7

# Rounding
puts f.floor
puts f.ceil
puts f.round
puts 3.14159.round

# Predicates
puts 0.0.zero?
puts 3.14.positive?
puts (-2.5).negative?

# Conversion
puts 3.5.to_i

#@ expect:
# 3
# 4
# 4
# 3
# true
# true
# true
# 3
