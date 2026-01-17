#@ run-pass
#@ check-output
# Scientific notation for float literals

# Basic scientific notation
x = 4.84e+00
y = 1.66e-03
z = 2.5E10

puts x
puts y
puts z

# Negative values with scientific notation
a = -3.14e+02
puts a

#@ expect:
# 4.84
# 0.00166
# 2.5e+10
# -314
