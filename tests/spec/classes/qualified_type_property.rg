#@ run-pass
#@ check-output
#
# Test qualified types in property declarations (e.g., big.Int, sync.Mutex)

import "sync"

class Counter
  property once: sync.Once
  property value: Int

  def initialize
    @value = 0
  end

  def increment_once
    @once.Do(-> { @value += 1 })
  end

  def get: Int
    @value
  end
end

def main
  c = Counter.new
  c.increment_once
  c.increment_once  # Should not increment again
  puts c.get
end

#@ expect:
# 1
