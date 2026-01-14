#@ compile-fail
#
# BUG (discovered): Range.include? returns Range instead of Bool
# Using it as if condition fails

if (1..10).include?(5)  #~ if condition must be Bool, got Range
  puts "includes 5"
end
