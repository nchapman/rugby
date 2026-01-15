#@ run-pass
#@ check-output
#
# Test: Section 17.1 - Goroutines

import "sync"

# go with expression
wg = sync.WaitGroup{}
wg.add(2)

go do
  puts "goroutine 1"
  wg.done
end

go do
  puts "goroutine 2"
  wg.done
end

wg.wait
puts "done"

#@ expect:
# goroutine 1
# goroutine 2
# done
