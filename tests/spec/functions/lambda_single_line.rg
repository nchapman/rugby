#@ run-pass
#@ check-output
#
# Test: Section 4.3/10.6 - Single-line lambdas with braces

# Basic lambda
double = -> (x : Int) { x * 2 }
puts double.(5)

# Lambda with multiple params
add = -> (a : Int, b : Int) { a + b }
puts add.(3, 4)

# Lambda in method call
nums = [1, 2, 3]
doubled = nums.map -> (n) { n * 2 }
puts doubled[0]
puts doubled[1]
puts doubled[2]

#@ expect:
# 10
# 7
# 2
# 4
# 6
