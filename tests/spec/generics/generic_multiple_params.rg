#@ run-pass
#@ check-output
#
# Test: Section 6.1 - Generic functions with multiple type parameters

def swap<T, U>(a: T, b: U): (U, T)
  b, a
end

x, y = swap(1, "hello")
puts x
puts y

#@ expect:
# hello
# 1
