#@ run-pass
#@ check-output
#
# Test: Section 19.1 - Array transformation methods

arr = [1, 2, 3, 4, 5]

# map
doubled = arr.map -> (x) { x * 2 }
puts doubled[0]
puts doubled[4]

# select / filter
evens = arr.select -> (x) { x % 2 == 0 }
puts evens.length
puts evens[0]

# reject
odds = arr.reject -> (x) { x % 2 == 0 }
puts odds.length

# compact (filter nil/zero)
mixed = [1, 0, 2, 0, 3]
compacted = mixed.compact
puts compacted.length

#@ expect:
# 2
# 10
# 2
# 2
# 3
# 3
