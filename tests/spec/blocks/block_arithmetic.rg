#@ run-pass
#@ check-output
#
# Test: Lambda parameter type inference
# Lambda variables should have proper types inferred from the iterable

nums = [1, 2, 3]
nums.each -> (n) do
  puts n * 10
end

# Works with strings too
words = ["hello", "world"]
words.each -> (w) do
  puts w.upcase
end

# Works with each_with_index
[10, 20, 30].each_with_index -> (val, i) do
  puts val + i
end

#@ expect:
# 10
# 20
# 30
# HELLO
# WORLD
# 10
# 21
# 32
