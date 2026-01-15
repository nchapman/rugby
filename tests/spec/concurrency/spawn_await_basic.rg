#@ run-pass
#@ check-output
#
# Test: Section 17.4 - Spawn and await

# Spawn returns Task<T>
t = spawn { 42 * 2 }

# Do other work
puts "working..."

# Await blocks until complete
result = await t
puts result

# Task with complex computation
t2 = spawn do
  sum = 0
  for i in 1..5
    sum = sum + i
  end
  sum
end

puts await t2

#@ expect:
# working...
# 84
# 15
