#@ run-pass
#@ check-output
#
# Test: Section 19.1 - Array iteration methods

arr = [1, 2, 3]

# each
arr.each -> (x) { puts x }

# each_with_index
arr.each_with_index -> (x, i) { puts "#{i}: #{x}" }

#@ expect:
# 1
# 2
# 3
# 0: 1
# 1: 2
# 2: 3
