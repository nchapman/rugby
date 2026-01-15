#@ run-pass
#@ check-output
#@ skip: Generics not yet implemented (Section 6.1)
#
# Test: Section 6.1 - Generic functions with multiple type parameters
# TODO: Implement generic functions

def swap<T, U>(a : T, b : U) -> (U, T)
  b, a
end

x, y = swap(1, "hello")
puts x
puts y

def map_values<K, V, R>(m : Map<K, V>, f : (V) -> R) -> Map<K, R>
  result = Map<K, R>{}
  for k, v in m
    result[k] = f.(v)
  end
  result
end

ages = {"alice" => 30, "bob" => 25}
doubled = map_values(ages, -> (v) { v * 2 })
puts doubled["alice"]

#@ expect:
# hello
# 1
# 60
