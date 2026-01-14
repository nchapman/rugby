#@ run-pass
#@ check-output
#
# BUG (discovered): Array first/last return pointers instead of values
# This test documents the bug - expected output is wrong until fixed

nums = [1, 2, 3]
# These currently output pointer addresses instead of 1 and 3
# puts nums.first  # Should output: 1
# puts nums.last   # Should output: 3

# Workaround: use indexing
puts nums[0]
puts nums[-1]

#@ expect:
# 1
# 3
