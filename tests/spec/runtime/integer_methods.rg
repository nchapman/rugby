#@ run-pass
#@ check-output
#
# Test: Section 19.4 - Integer methods

n = 42

# Predicates
puts n.even?
puts n.odd?
puts n.zero?
puts 0.zero?
puts n.positive?
puts (-5).negative?

# Math
puts (-10).abs
puts 5.clamp(1, 3)
puts 2.clamp(1, 3)

# Iteration
3.times -> (i) { puts i }

# Conversion
puts n.to_s

#@ expect:
# true
# false
# false
# true
# true
# true
# 10
# 3
# 2
# 0
# 1
# 2
# 42
