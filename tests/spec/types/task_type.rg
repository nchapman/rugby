#@ run-pass
#@ check-output
#
# Test: Section 5.12 - Task type

# Spawn returns Task<T>
t = spawn { 42 }

# Await gets the result
result = await t
puts result

# Task with computation
t2 = spawn do
  sum = 0
  for i in 1..10
    sum = sum + i
  end
  sum
end

puts await t2

#@ expect:
# 42
# 55
