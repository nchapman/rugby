#@ run-pass
#@ check-output
#
# Test: Section 5.10 - Range type

# Inclusive range
r1 = 1..5
puts r1.include?(1)
puts r1.include?(5)
puts r1.include?(0)

# Exclusive range
r2 = 1...5
puts r2.include?(1)
puts r2.include?(4)
puts r2.include?(5)

# Range to array
arr = (1..3).to_a
puts arr.length
puts arr[0]
puts arr[1]
puts arr[2]

# Range iteration
for i in 1..3
  puts i
end

#@ expect:
# true
# true
# false
# true
# true
# false
# 3
# 1
# 2
# 3
# 1
# 2
# 3
