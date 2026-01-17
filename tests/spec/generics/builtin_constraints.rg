#@ run-pass
#@ check-output
#
# Test: Section 6.4 - Built-in type constraints
# Tests Numeric, Ordered, and Equatable constraints with generic type parameters

# Numeric constraint with T.zero
def sum<T : Numeric>(values : Array<T>) -> T
  values.reduce(T.zero) -> { |acc, v| acc + v }
end

nums = [1, 2, 3, 4, 5]
puts sum(nums)

# Ordered constraint
def max<T : Ordered>(a : T, b : T) -> T
  a > b ? a : b
end

puts max(10, 20)
puts max("apple", "banana")

# Equatable constraint
def contains<T : Equatable>(arr : Array<T>, val : T) -> Bool
  arr.any? -> { |x| x == val }
end

puts contains([1, 2, 3], 2)

#@ expect:
# 15
# 20
# banana
# true
