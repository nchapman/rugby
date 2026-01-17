#@ run-pass
#@ check-output
# Multi-return reassignment inside blocks works
# Reassigning existing variables with multi-return inside if/block
# correctly uses = (assignment) instead of := (declaration)

import "strconv"

n = 4

# Reassignment works inside blocks
if true
  n, _ = strconv.atoi("10")
end

puts n

# Also works in while loops
m = 0
i = 0
while i < 3
  m, _ = strconv.atoi("5")
  i += 1
end

puts m

#@ expect:
# 10
# 5
