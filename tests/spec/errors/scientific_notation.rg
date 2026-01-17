#@ compile-fail
# Bug: No scientific notation for floats
# Float literals with e+NN or e-NN notation fail to parse

# Expected behavior:
# x = 4.84143144246472090e+00  # should work
# y = 1.66007664274403694e-03  # should work

# Current: fails to parse scientific notation
x = 4.84e+00
#~ ERROR: expected
