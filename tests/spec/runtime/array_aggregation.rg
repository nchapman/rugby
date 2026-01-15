#@ run-pass
#@ check-output
#
# Test: Section 19.1 - Array aggregation methods

arr = [1, 2, 3, 4, 5]

# reduce
sum = arr.reduce(0) -> (acc, x) { acc + x }
puts sum

product = arr.reduce(1) -> (acc, x) { acc * x }
puts product

# sum (numeric arrays)
puts arr.sum

# min / max
puts arr.min ?? 0
puts arr.max ?? 0

#@ expect:
# 15
# 120
# 15
# 1
# 5
