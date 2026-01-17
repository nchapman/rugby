#@ run-pass
#@ check-output
# Bitwise shift operators

# Left shift
a = 1 << 4
puts a

# Right shift
b = 16 >> 2
puts b

# Chained shifts
c = 1 << 3 >> 1
puts c

# With variables
n = 8
shift = 2
d = n << shift
e = n >> shift
puts d
puts e

#@ expect:
# 16
# 4
# 4
# 32
# 2
