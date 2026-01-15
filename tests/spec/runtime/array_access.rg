#@ run-pass
#@ check-output
#
# Test: Section 19.1 - Array access methods

arr = [1, 2, 3, 4, 5]

# first / last
puts arr.first ?? 0
puts arr.last ?? 0

# take / drop
puts arr.take(2).length
puts arr.drop(3).length

# length / size
puts arr.length
puts arr.size

# empty?
puts arr.empty?
puts [].empty?

#@ expect:
# 1
# 5
# 2
# 2
# 5
# 5
# false
# true
