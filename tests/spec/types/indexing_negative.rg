#@ run-pass
#@ check-output
#
# Test: Section 5.6 - Negative indexing

arr = ["a", "b", "c", "d", "e"]

# Negative indices count from end
puts arr[-1]
puts arr[-2]
puts arr[-3]

# String indexing
str = "hello"
puts str[-1]
puts str[-2]

#@ expect:
# e
# d
# c
# o
# l
