#@ run-pass
#@ check-output

# Test: Array splat operator in literals (Section 5.5)
# [1, *rest, 4] expands rest array into the literal

# Basic splat in middle
middle = [2, 3]
arr1 = [1, *middle, 4]
puts arr1.length
puts arr1[0]
puts arr1[1]
puts arr1[2]
puts arr1[3]

# Splat at start
suffix = [3, 4, 5]
arr2 = [*suffix, 6]
puts arr2.length
puts arr2[0]

# Splat at end
prefix = [1, 2]
arr3 = [0, *prefix]
puts arr3.length
puts arr3[2]

# Multiple splats
a = [1, 2]
b = [3, 4]
arr4 = [*a, *b]
puts arr4.length

# Empty array splat
empty : Array<Int> = []
arr5 = [1, *empty, 2]
puts arr5.length
puts arr5[0]
puts arr5[1]

# Splat with single element
single = [100]
arr6 = [*single]
puts arr6.length
puts arr6[0]

#@ expect:
# 4
# 1
# 2
# 3
# 4
# 4
# 3
# 3
# 2
# 4
# 2
# 1
# 2
# 1
# 100
