#@ run-pass
#@ check-output
#
# Test: Range methods coverage

# Inclusive range
r1 = 1..5
puts r1.size
puts r1.include?(3)
puts r1.include?(5)
puts r1.include?(6)

# Exclusive range
r2 = 1...5
puts r2.size
puts r2.include?(4)
puts r2.include?(5)

# Range to_a
arr = (1..3).to_a
puts arr.length
puts arr[0]
puts arr[2]

# Range each
sum = 0
(1..5).each do |n|
  sum += n
end
puts sum

# Range with negative numbers
neg = -2..2
puts neg.size
puts neg.include?(0)
puts neg.include?(-2)

# Range to_a with negative
neg_arr = (-1..1).to_a
puts neg_arr.length
puts neg_arr[0]
puts neg_arr[2]

#@ expect:
# 5
# true
# true
# false
# 4
# true
# false
# 3
# 1
# 3
# 15
# 5
# true
# true
# 3
# -1
# 1
