#@ compile-fail
# Bug: No make() for pre-sized arrays
# Go has make([]T, size) to allocate arrays with a given size.
# Ruby has Array.new(n, default).
# Rugby has no equivalent.

# Expected behavior:
# flags = Array<Bool>.new(1000, false)  # pre-allocate 1000 falses
# OR
# flags = make(Array<Bool>, 1000)

# Current workaround is slow for benchmarks:
# n = 1000
# flags : Array<Bool> = []
# i = 0
# while i < n
#   flags << false
#   i += 1
# end

# This should create a pre-sized array but fails:
flags = Array<Bool>.new(100, false)
#~ ERROR: undefined
