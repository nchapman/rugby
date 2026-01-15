#@ run-pass
#@ check-output
#
# Test: Section 9.1 - Variables

# Basic assignment
x = 42
puts x

# Explicit type annotation
y : Int = 100
puts y

# Compound operators
x += 5
puts x

x -= 3
puts x

x *= 2
puts x

x /= 4
puts x

# Reassignment
x = 10
puts x

#@ expect:
# 42
# 100
# 47
# 44
# 88
# 22
# 10
