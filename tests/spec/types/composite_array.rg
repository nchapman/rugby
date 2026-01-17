#@ run-pass
#@ check-output
#
# Test: Section 5.2/5.5 - Array<T> composite type

# Array literals with type inference
nums = [1, 2, 3]
puts nums.length
puts nums[0]
puts nums[1]
puts nums[2]

# String array
words = ["hello", "world"]
puts words[0]
puts words[1]

# Empty typed array
empty: Array<Int> = []
puts empty.length

# Append with <<
arr = [1, 2]
arr << 3
puts arr.length
puts arr[2]

#@ expect:
# 3
# 1
# 2
# 3
# hello
# world
# 0
# 3
# 3
