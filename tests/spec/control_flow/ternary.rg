#@ run-pass
#@ check-output

# Test: Ternary operator (Section 9.4)
# The ternary operator is right-associative: a > b ? a : b

# Basic ternary
a = 10
b = 5
max = a > b ? a : b
puts max

# Ternary with function calls
def get_a: Int
  100
end

def get_b: Int
  50
end

result = get_a() > get_b() ? get_a() : get_b()
puts result

# Nested ternary (right-associative)
x = 15
category = x > 20 ? "large" : x > 10 ? "medium" : "small"
puts category

# Ternary in assignment
y = 7
status = y.even? ? "even" : "odd"
puts status

# Ternary with boolean result
check = true ? "yes" : "no"
puts check

# Ternary with expressions
sum = (a > b ? a : b) + 5
puts sum

#@ expect:
# 10
# 100
# medium
# odd
# yes
# 15
