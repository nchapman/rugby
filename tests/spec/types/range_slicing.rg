#@ run-pass
#@ check-output
#
# Test: Section 5.7 - Range slicing

arr = ["a", "b", "c", "d", "e"]

# Inclusive range slice
slice1 = arr[1..3]
puts slice1.length
puts slice1[0]
puts slice1[1]
puts slice1[2]

# Exclusive range slice
slice2 = arr[1...3]
puts slice2.length
puts slice2[0]
puts slice2[1]

# String slicing
str = "hello"
puts str[1..3]

#@ expect:
# 3
# b
# c
# d
# 2
# b
# c
# ell
