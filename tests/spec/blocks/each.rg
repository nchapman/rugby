#@ run-pass
#@ check-output
#
# Test: Lambda iteration with each

nums = [1, 2, 3]

nums.each -> { |n| puts n }

#@ expect:
# 1
# 2
# 3
