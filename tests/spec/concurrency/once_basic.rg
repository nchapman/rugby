#@ run-pass
#@ check-output
#
# Test: Section 17.6 - Once (executed exactly once)

import "sync"

once = sync.Once{}
initialized = false

wg = sync.WaitGroup{}
wg.Add(3)

for i in 1..3
  go do
    once.Do -> do
      puts "initializing"
      initialized = true
    end
    wg.Done
  end
end

wg.Wait
puts initialized

#@ expect:
# initializing
# true
