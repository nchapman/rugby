#@ run-pass
#@ check-output
#
# Test spawn and await for async tasks

# Basic spawn and await
t = spawn { 42 }
result = await t
puts result

# Spawn with computation
t2 = spawn do
  x = 10
  y = 20
  x + y
end
result2 = await t2
puts result2

# TODO: Spawn capturing outer variable doesn't work
#       outer = 100
#       t3 = spawn { outer + 5 }
#       puts await t3  # fails with undefined: puts

#@ expect:
# 42
# 30
