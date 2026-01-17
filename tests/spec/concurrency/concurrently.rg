#@ run-pass
#@ check-output
#
# Test concurrently blocks for structured concurrency
#
# Concurrently blocks ensure all spawned tasks complete before
# the block exits, providing cleanup guarantees.

# Simple concurrently block
result = concurrently do |scope|
  a = scope.spawn -> { 10 }
  b = scope.spawn -> { 20 }

  await(a) + await(b)
end

puts result

#@ expect:
# 30
