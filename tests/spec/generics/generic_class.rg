#@ compile-fail
#@ skip: Generics not yet implemented (Section 6.2)
#
# Test: Section 6.2 - Generic classes
# TODO: Implement generic class syntax

class Box<T>
  def initialize(@value : T)
  end

  def get -> T
    @value
  end

  def map<R>(f : (T) -> R) -> Box<R>
    Box<R>.new(f.(@value))
  end
end

box = Box<Int>.new(42)
puts box.get
