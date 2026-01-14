#@ run-fail
#
# Test: Runtime panic on array index out of bounds
# This compiles successfully but fails at runtime

arr = [1, 2, 3]
puts arr[99]  #~ ERROR: index out of range
