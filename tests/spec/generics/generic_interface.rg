#@ compile-fail
#@ skip: Generics not yet implemented (Section 6.3)
#
# Test: Section 6.3 - Generic interfaces
# TODO: Implement generic interface syntax

interface Container<T>
  def get -> T?
  def put(value : T)
end
