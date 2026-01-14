#@ run-pass
#@ check-output
#
# Test spawn blocks capturing outer variables
#
# Spawn blocks should be able to capture and use variables from
# the enclosing scope, just like regular closures in Ruby.

# Simple capture of outer variables
x = 10
y = 20

task = spawn do
  x + y
end

result = task.await
puts result

# Capture block parameter
multiplier = 5
task2 = spawn do
  multiplier * 2
end

puts task2.await

#@ expect:
# 30
# 10
