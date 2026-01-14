#@ run-pass
#@ check-output
#
# Test: Reduce block method

nums = [1, 2, 3, 4, 5]

sum = nums.reduce(0) { |acc, n| acc + n }
puts sum

product = nums.reduce(1) { |acc, n| acc * n }
puts product

#@ expect:
# 15
# 120
