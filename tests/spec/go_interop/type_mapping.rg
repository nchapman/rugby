#@ run-pass
#@ check-output
#
# Test: Section 18.3 - Type mapping between Rugby and Go

# Int -> int
x: Int = 42
puts x

# String -> string
s: String = "hello"
puts s

# Array<T> -> []T
arr: Array<Int> = [1, 2, 3]
puts arr.length

# Map<K, V> -> map[K]V
m: Map<String, Int> = {"a" => 1}
puts m["a"]

# Bytes -> []byte
b = "hello".bytes
puts b.length

#@ expect:
# 42
# hello
# 3
# 1
# 5
