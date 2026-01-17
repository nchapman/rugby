#@ run-pass
#@ check-output
#
# Test that empty array literals can be assigned to typed arrays

empty_nums: Array<Int> = []
empty_strings: Array<String> = []

# Verify the arrays are empty
puts empty_nums.size
puts empty_strings.size

# Verify types compile correctly
nums: Array<Int> = [1, 2, 3]
puts nums.size

#@ expect:
# 0
# 0
# 3
