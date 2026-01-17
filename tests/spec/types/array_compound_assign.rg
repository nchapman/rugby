#@ run-pass
#@ check-output
# Compound assignment on array indices
# arr[i] += value etc works correctly

arr = [10, 20, 30]

# Test all compound operators
arr[0] += 5
arr[1] -= 10
arr[2] *= 2

puts arr[0]
puts arr[1]
puts arr[2]

# Test with variable index
i = 0
arr[i] += 100

puts arr[i]

# Test with map
counts = {"a" => 5, "b" => 10}
counts["a"] += 2
counts["b"] -= 3

puts counts["a"]
puts counts["b"]

#@ expect:
# 15
# 10
# 60
# 115
# 7
# 7
