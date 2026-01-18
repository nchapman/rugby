#@ run-pass
#@ check-output

# Test that compound assignment doesn't double-evaluate expressions with side effects

class CallTracker
  property count: Int

  def initialize
    @count = 0
  end
end

class Counter
  def initialize(@value: Int)
  end

  def value: Int
    @value
  end

  def value=(v: Int)
    @value = v
  end
end

tracker = CallTracker.new

def get_counter(t: CallTracker): Counter
  t.count += 1
  Counter.new(10)
end

# Simple case - identifier (no temp var needed)
counter = Counter.new(5)
counter.value += 3
puts counter.value

# Complex case - function call (needs temp var to avoid double evaluation)
tracker.count = 0
get_counter(tracker).value += 5
puts tracker.count  # Should be 1, not 2

# Chained selector - should be safe
counter2 = Counter.new(20)
counter2.value -= 5
puts counter2.value

#@ expect:
# 8
# 1
# 15
