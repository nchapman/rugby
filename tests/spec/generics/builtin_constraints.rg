#@ run-pass
#@ check-output
#@ skip: TODO: Implement T.zero for Numeric constraint
#
# Test: Section 6.4 - Built-in type constraints
# TODO: Implement generic constraints

# Numeric constraint
def sum<T : Numeric>(values : Array<T>) -> T
  values.reduce(T.zero) -> (acc, v) { acc + v }
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
  arr.any? -> (x) { x == val }
end

puts contains([1, 2, 3], 2)

# Hashable constraint
def unique_keys<K : Hashable, V>(pairs : Array<(K, V)>) -> Array<K>
  seen = Set<K>{}
  result = Array<K>{}
  for k, _ in pairs
    unless seen.include?(k)
      seen.add(k)
      result << k
    end
  end
  result
end

#@ expect:
# 15
# 20
# banana
# true
