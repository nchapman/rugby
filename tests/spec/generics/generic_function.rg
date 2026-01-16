#@ run-pass
#@ check-output
#
# Test: Section 6.1 - Generic functions

def identity<T>(x : T) -> T
  x
end

# Type parameters inferred from arguments
puts identity(42)
puts identity("hello")

#@ expect:
# 42
# hello
