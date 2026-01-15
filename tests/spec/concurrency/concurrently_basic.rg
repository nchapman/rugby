#@ run-pass
#@ check-output
#
# Test: Section 17.5 - Structured concurrency with concurrently

def compute_a -> Int
  10
end

def compute_b -> Int
  20
end

result = concurrently -> (scope) do
  a = scope.spawn { compute_a }
  b = scope.spawn { compute_b }

  x = await a
  y = await b
  x + y
end

puts result

#@ expect:
# 30
