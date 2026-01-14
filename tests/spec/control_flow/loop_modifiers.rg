#@ run-pass
#@ check-output
#
# Test loop modifiers (while/until at end of line)
# Loop modifiers execute the statement on the left repeatedly
# TODO: Should support compound assignment in loop modifier expressions
#       e.g., `puts counter += 1 while counter < 3`

# Simple while modifier - print and remove from array
items = ["a", "b", "c"]
puts items.shift while items.length > 0

# Simple until modifier
arr = [1, 2, 3]
puts arr.shift until arr.empty?

#@ expect:
# a
# b
# c
# 1
# 2
# 3
