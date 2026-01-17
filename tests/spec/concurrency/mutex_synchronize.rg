#@ skip: Mutex.synchronize is not yet implemented (spec feature)
#@ run-pass
#@ check-output

# Test: Mutex.synchronize block pattern (Section 17.6)
# synchronize takes a lambda and handles lock/unlock automatically

import "sync"

mutex = sync.Mutex.new
counter = 0

# Basic synchronize pattern
mutex.synchronize -> do
  counter += 1
end

puts counter

# Multiple synchronized blocks
5.times -> do
  mutex.synchronize -> do
    counter += 1
  end
end

puts counter

# Synchronize with result value
result = 0
mutex.synchronize -> do
  result = counter * 2
end
puts result

# Nested data structure modification
data : Array<Int> = []
mutex.synchronize -> do
  data << 10
  data << 20
  data << 30
end
puts data.length

#@ expect:
# 1
# 6
# 12
# 3
