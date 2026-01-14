#@ run-pass
#@ check-output
#
# Test: Range literals and operations

# Inclusive range
r = 1..5
puts r.size

# For-in with range
for i in 1..3
  puts i
end

#@ expect:
# 5
# 1
# 2
# 3
