#@ run-pass
#@ check-output
#
# Test: Section 19.1 - Array search methods

arr = [1, 2, 3, 4, 5]

# find / detect
found = arr.find -> (x) { x > 3 }
puts found ?? 0

# any?
puts arr.any? -> (x) { x > 4 }
puts arr.any? -> (x) { x > 10 }

# all?
puts arr.all? -> (x) { x > 0 }
puts arr.all? -> (x) { x > 2 }

# none?
puts arr.none? -> (x) { x < 0 }

# include? / contains?
puts arr.include?(3)
puts arr.include?(10)

#@ expect:
# 4
# true
# false
# true
# false
# true
# true
# false
