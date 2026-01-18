#@ run-pass
#@ check-output
#
# Test: Extended Float methods coverage

# abs
puts (-3.5).abs
puts 3.5.abs

# truncate
puts 3.7.truncate
puts (-3.7).truncate

# floor and ceil edge cases
puts (-3.2).floor
puts (-3.2).ceil

# to_s
puts 3.14.to_s

# predicates with edge cases
puts 0.0.positive?
puts 0.0.negative?
puts 0.001.positive?
puts (-0.001).negative?

# round (Go uses round half away from zero)
puts 2.5.round
puts 3.5.round
puts 2.4.round
puts 2.6.round

#@ expect:
# 3.5
# 3.5
# 3
# -3
# -4
# -3
# 3.14
# false
# false
# true
# true
# 3
# 4
# 2
# 3
