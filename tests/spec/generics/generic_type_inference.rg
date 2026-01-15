#@ run-pass
#@ check-output
#@ skip: Generics not yet implemented (Section 6.5)
#
# Test: Section 6.5 - Type inference for generics
# TODO: Implement generic type inference

# Type parameters inferred from arguments
def identity<T>(x : T) -> T
  x
end

# Inferred as Int
result1 = identity(42)
puts result1

# Inferred as String
result2 = identity("hello")
puts result2

# Explicit type parameter when needed
box = Box<Int>.new(42)

# Empty array needs explicit type
result : Array<String> = []

#@ expect:
# 42
# hello
