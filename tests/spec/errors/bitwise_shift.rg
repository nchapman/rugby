#@ compile-fail
# Bug: No bitwise shift operators
# The << operator is only for array append and channel send.
# There's no bitwise left/right shift.

# Expected behavior:
# n = 1 << 4   # should be 16 (bitwise shift left)
# m = 16 >> 2  # should be 4 (bitwise shift right)

# Current: << only works for array/channel append
arr : Array<Int> = []
arr << 1

# This should be bitwise shift but fails:
n = 1 << 4
#~ ERROR: expected
