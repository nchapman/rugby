#@ compile-fail
#
# Test: Section 12.4 - Struct immutability compile error
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
