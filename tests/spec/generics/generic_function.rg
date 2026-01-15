#@ compile-fail
#@ skip: Generics not yet implemented (Section 6.1)
#
# Test: Section 6.1 - Generic functions
# TODO: Implement generic function syntax

def identity<T>(x : T) -> T
  x
end

# Type parameters inferred from arguments
puts identity(42)
puts identity("hello")
