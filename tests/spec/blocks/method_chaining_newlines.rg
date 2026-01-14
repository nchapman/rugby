#@ compile-fail
#
# BUG-038: Method chaining with newlines
# Method chaining that spans multiple lines should work but fails to parse

result = [1, 2, 3, 4, 5, 6]
  .select { |n| n.even? }  #~ expected 'end'
  .map { |n| n * 10 }
