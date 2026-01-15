#@ run-pass
#@ check-output
#
# Test: Section 17.6 - Mutex synchronization

import "sync"

mutex = sync.Mutex{}
counter = 0

wg = sync.WaitGroup{}
wg.Add(3)

for i in 1..3
  go do
    mutex.Lock
    counter += 10
    mutex.Unlock
    wg.Done
  end
end

wg.Wait
puts counter

#@ expect:
# 30
