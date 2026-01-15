#@ run-pass
#@ check-output
#
# Test goroutines with go keyword

import "sync"

# Use a WaitGroup to wait for goroutines
wg = sync.WaitGroup.new
ch = Chan<String>.new(3)

wg.add(3)

go do
  ch << "first"
  wg.done
end

go do
  ch << "second"
  wg.done
end

go do
  ch << "third"
  wg.done
end

wg.wait

# Collect results (order may vary, so we collect and sort)
results : Array<String> = []
results = results << ch.receive
results = results << ch.receive
results = results << ch.receive

# Print sorted for deterministic output
results = results.sorted
results.each -> (r) { puts r }

#@ expect:
# first
# second
# third
