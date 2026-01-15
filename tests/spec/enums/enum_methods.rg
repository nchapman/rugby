#@ compile-fail
#@ skip: Enums not yet implemented (Section 7.3)
#
# Test: Section 7.3 - Enum methods
# TODO: Implement enum methods

enum Color
  Red
  Green
  Blue
end

puts Color::Red.to_s
puts Color.values.length
