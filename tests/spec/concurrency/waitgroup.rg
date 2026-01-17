#@ run-pass
#@ check-output
#
# Test WaitGroup usage via Go interop
#
# WaitGroup is available via sync package for explicit concurrency control.

import "sync"

wg = sync.WaitGroup.new

count = 0

3.times -> do |i|
  wg.add(1)
  go do
    sleep 0.01
    count += 1
    wg.done
  end
end

wg.wait
puts count

#@ expect:
# 3
