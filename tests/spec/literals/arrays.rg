#@ run-pass
#@ check-output
#
# Test: Array literals and operations

# Basic array
nums = [1, 2, 3]
puts nums[0]
puts nums[1]
puts nums[2]

# Negative indexing
puts nums[-1]

# String array
words = ["one", "two", "three"]
puts words[1]

# Array methods
puts nums.length

#@ expect:
# 1
# 2
# 3
# 3
# two
# 3
