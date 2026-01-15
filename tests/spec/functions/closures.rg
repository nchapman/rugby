#@ run-pass
#@ check-output
#
# Test: Section 10.7 - Closures capture variables

def make_counter -> () -> Int
  count = 0
  -> { count += 1 }
end

counter = make_counter
puts counter.()
puts counter.()
puts counter.()

# Capture by reference
multiplier = 2
scale = -> (x : Int) { x * multiplier }
puts scale.(5)

multiplier = 3
puts scale.(5)

#@ expect:
# 1
# 2
# 3
# 10
# 15
