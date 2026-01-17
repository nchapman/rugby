#@ run-pass
#@ check-output
# Pre-sized array allocation with Array<T>.new(size, default)

# Create array of 5 booleans, all false
flags = Array<Bool>.new(5, false)
puts flags.size
puts flags[0]
puts flags[4]

# Create array of 3 integers, all 42
nums = Array<Int>.new(3, 42)
puts nums.size
puts nums[0]
puts nums[1]
puts nums[2]

# Create array of 2 strings
names = Array<String>.new(2, "unknown")
puts names.size
puts names[0]

# Zero-sized array
empty = Array<Int>.new(0, 0)
puts empty.size

# Size-only allocation (zero values)
zeros = Array<Int>.new(3)
puts zeros.size
puts zeros[0]

#@ expect:
# 5
# false
# false
# 3
# 42
# 42
# 42
# 2
# unknown
# 0
# 3
# 0
