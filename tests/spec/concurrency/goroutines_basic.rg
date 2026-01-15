#@ run-pass
#@ check-output
#
# Test: Section 17.1 - Goroutines

import "sync"

# go with expression
wg = sync.WaitGroup{}
wg.Add(2)

go do
  puts "goroutine 1"
  wg.Done
end

go do
  puts "goroutine 2"
  wg.Done
end

wg.Wait
puts "done"

#@ expect:
# goroutine 1
# goroutine 2
# done
