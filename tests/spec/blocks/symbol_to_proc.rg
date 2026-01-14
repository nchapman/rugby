#@ run-pass
#@ check-output
#
# Test symbol-to-proc syntax (&:method)
#
# Symbol-to-proc converts a symbol to a block that calls that method
# on each element. It's syntactic sugar for common block patterns.
#
# Example: names.map(&:upcase) is equivalent to names.map { |x| x.upcase }

# Basic usage with map
names = ["hello", "world"]
upper = names.map(&:upcase)
puts upper[0]
puts upper[1]

# With select (predicate method)
words = ["", "hello", "", "world", ""]
non_empty = words.select(&:present?)
puts non_empty.length

# With integers
numbers = [1, -2, 3, -4, 5]
absolute = numbers.map(&:abs)
puts absolute[1]
puts absolute[3]

#@ expect:
# HELLO
# WORLD
# 2
# 2
# 4
