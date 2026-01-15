#@ run-pass
#@ check-output
#
# Test: Section 17.6 - Mutex synchronization

import "sync"

mutex = sync.Mutex{}
counter = 0

wg = sync.WaitGroup{}
wg.add(3)

for i in 1..3
  go do
    mutex.lock
    counter += 10
    mutex.unlock
    wg.done
  end
end

wg.wait
puts counter

#@ expect:
# 30
