#@ run-pass
#@ check-output
#
# Test that method chaining works with newlines

result = [1, 2, 3, 4, 5, 6]
  .select { |n| n.even? }
  .map { |n| n * 10 }

result.each { |n| puts n }

#@ expect:
# 20
# 40
# 60
