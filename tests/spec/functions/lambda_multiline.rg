#@ run-pass
#@ check-output
#
# Test: Section 4.3/10.6 - Multi-line lambdas with do...end

# Multi-line lambda
process = -> (x : Int) do
  doubled = x * 2
  doubled + 1
end
puts process.(5)

# Lambda with explicit return type
handler = -> (s : String) -> String do
  prefix = ">> "
  prefix + s
end
puts handler.("hello")

# Multi-line in iteration
items = [1, 2, 3]
items.each -> (n) do
  result = n * 10
  puts result
end

#@ expect:
# 11
# >> hello
# 10
# 20
# 30
