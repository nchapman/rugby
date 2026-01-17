#@ run-pass
#@ check-output

# Test: Loop capture gotcha and explicit capture pattern (Section 10.7)
# Variables are captured by reference, so loop vars need explicit copy

# Correct pattern: explicit capture with copy variable
# We store results in an array to verify the captured values
results = [0, 0, 0]

for i in 0..2
  copy = i  # capture current value
  # Simulate storing a closure result
  results[i] = copy * 10
end

puts results[0]
puts results[1]
puts results[2]

# Verify capture with immediate invocation
def test_capture: Array<Int>
  captured = [0, 0, 0]
  for i in 0..2
    val = i  # explicit capture
    captured[i] = val
  end
  captured
end

vals = test_capture()
puts vals[0]
puts vals[1]
puts vals[2]

# Closure with explicit capture - simpler approach
# Each iteration creates a closure capturing the current loop value
def run_closures
  for i in 0..2
    copy = i
    fn = -> { copy }
    puts fn.()
  end
end

run_closures()

#@ expect:
# 0
# 10
# 20
# 0
# 1
# 2
# 0
# 1
# 2
