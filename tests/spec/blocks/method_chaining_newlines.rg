#@ run-pass
#@ check-output
#
# Test that method chaining works with lambdas

# Use parenthesized form for chaining
result = [1, 2, 3, 4, 5, 6].select(-> { |n| n.even? }).map(-> { |n| n * 10 })

result.each -> { |n| puts n }

#@ expect:
# 20
# 40
# 60
