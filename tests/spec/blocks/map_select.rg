#@ run-pass
#@ check-output
#
# Test: Map and select block methods

nums = [1, 2, 3, 4, 5]

# Map
doubled = nums.map { |n| n * 2 }
doubled.each { |n| puts n }

# Select
evens = nums.select { |n| n.even? }
evens.each { |n| puts n }

# Reject
odds = nums.reject { |n| n.even? }
odds.each { |n| puts n }

#@ expect:
# 2
# 4
# 6
# 8
# 10
# 2
# 4
# 1
# 3
# 5
