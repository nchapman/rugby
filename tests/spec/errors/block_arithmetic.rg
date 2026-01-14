#@ compile-fail
#
# BUG (discovered): Block variables are typed as 'any' instead of inferred element type
# Arithmetic on block variables fails

nums = [1, 2, 3]
nums.each do |n|
  puts n * 10  #~ mismatched types any and int
end
