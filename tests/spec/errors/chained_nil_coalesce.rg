#@ compile-fail
# Bug: Chained ?? operators don't work with optional fallbacks
# The ?? operator requires right side to be T, not T?
# This prevents chaining like: a ?? b ?? c

a : Int? = nil
b : Int? = 50
c : Int? = nil

# Expected behavior:
# result = a ?? b ?? c  # should return 50

# Current: fails because b is Int?, not Int
result = a ?? b ?? c
puts result ?? -1
#~ ERROR: type mismatch
