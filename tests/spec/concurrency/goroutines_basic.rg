#@ run-pass
#@ check-output
#
# Test: Section 17.1 - Goroutines

import "sync"

# go with expression
wg = sync.WaitGroup{}
wg.add(1)

go do
  puts "goroutine"
  wg.done
end

wg.wait
puts "done"

#@ expect:
# goroutine
# done
