#@ compile-fail
#@ skip: Structs not yet implemented (Section 12.4)
#
# Test: Section 12.4 - Struct immutability compile error
# TODO: Implement struct immutability enforcement
#
# This test should fail to compile because structs are immutable

struct Point
  x : Int
  y : Int

  def try_move(dx : Int)
    @x += dx  # ERROR: cannot modify struct field
  end
end

#~ ERROR: cannot modify struct field
