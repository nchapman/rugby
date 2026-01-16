#@ skip: Lambda return type mismatch with sync.Once.Do - need type inference for Go interop
#
# Test: Section 17.6 - Once (executed exactly once)

import "sync"

once = sync.Once{}
initialized = false

wg = sync.WaitGroup{}
wg.add(3)

for i in 1..3
  go do
    once.do -> do
      puts "initializing"
      initialized = true
    end
    wg.done
  end
end

wg.wait
puts initialized

#@ expect:
# initializing
# true
