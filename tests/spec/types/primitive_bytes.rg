#@ run-pass
#@ check-output
#
# Test: Section 5.1 - Bytes type (mutable byte slice)

# Create from string
b = "hello".bytes
puts b.length

# Access individual bytes
puts b[0]
puts b[1]

#@ expect:
# 5
# 104
# 101
