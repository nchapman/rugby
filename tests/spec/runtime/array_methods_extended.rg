#@ run-pass
#@ check-output
#
# Test: Extended Array methods coverage

# flatten
nested = [[1, 2], [3, 4], [5]]
flat = nested.flatten
puts flat.length
puts flat[0]
puts flat[4]

# uniq
with_dups = [1, 2, 2, 3, 3, 3, 4]
unique = with_dups.uniq
puts unique.length

# min and max
nums = [3, 1, 4, 1, 5, 9, 2, 6]
puts nums.min
puts nums.max

# sum
puts [1, 2, 3, 4, 5].sum

# first and last
items = [10, 20, 30, 40, 50]
puts items.first
puts items.last

# take and drop
puts [1, 2, 3, 4, 5].take(3).length
puts [1, 2, 3, 4, 5].drop(2).length

# reversed (non-mutating)
original = [1, 2, 3]
rev = original.reversed
puts rev[0]
puts original[0]

# sorted
unsorted = [3, 1, 4, 1, 5]
sorted = unsorted.sorted
puts sorted[0]
puts sorted[4]

# join
puts ["a", "b", "c"].join("-")

# contains?
puts [1, 2, 3].contains?(2)
puts [1, 2, 3].contains?(5)

# empty?
puts [].empty?
puts [1].empty?

#@ expect:
# 5
# 1
# 5
# 4
# 1
# 9
# 15
# 10
# 50
# 3
# 3
# 3
# 1
# 1
# 5
# a-b-c
# true
# false
# true
# false
