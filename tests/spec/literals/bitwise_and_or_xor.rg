#@ run-pass
#@ check-output
# Bitwise AND, OR, XOR operators

# Bitwise AND
a = 0b1100 & 0b1010
puts a  # 0b1000 = 8

# Bitwise OR
b = 0b1100 | 0b1010
puts b  # 0b1110 = 14

# Bitwise XOR
c = 0b1100 ^ 0b1010
puts c  # 0b0110 = 6

# Combined with shift
d = (1 << 4) & 0xFF
puts d  # 16

# Masking a byte
val = 0xABCD
masked = (val >> 4) & 0xF
puts masked  # 0xC = 12

# XOR swap pattern
x = 5
y = 3
x = x ^ y
y = x ^ y
x = x ^ y
puts x
puts y

#@ expect:
# 8
# 14
# 6
# 16
# 12
# 3
# 5
