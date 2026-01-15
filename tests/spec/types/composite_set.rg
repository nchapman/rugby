#@ run-pass
#@ check-output
#
# Test: Section 5.2/5.9 - Set<T> composite type

# Set literal
s = Set{1, 2, 3}
puts s.length
puts s.include?(1)
puts s.include?(5)

# Typed set
s2 = Set<String>{"a", "b"}
puts s2.length
puts s2.include?("a")

# Set operations
a = Set{1, 2, 3}
b = Set{2, 3, 4}

# Union
union = a | b
puts union.length

# Intersection
inter = a & b
puts inter.length

# Difference
diff = a - b
puts diff.length

#@ expect:
# 3
# true
# false
# 2
# true
# 4
# 2
# 1
