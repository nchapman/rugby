#@ run-pass
#@ check-output
#
# Test goroutines with go keyword

import sync

# Use a WaitGroup to wait for goroutines
wg = sync.WaitGroup.new
ch = Chan[String].new(3)

wg.Add(3)

go do
  ch << "first"
  wg.Done
end

go do
  ch << "second"
  wg.Done
end

go do
  ch << "third"
  wg.Done
end

wg.Wait

# Collect results (order may vary, so we collect and sort)
results : Array[String] = []
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
