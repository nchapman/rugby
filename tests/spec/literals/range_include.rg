#@ run-pass
#@ check-output
#
# Test: Range.include? method

if (1..10).include?(5)
  puts "includes 5"
end

# Also works with exclusive ranges
if !(1...5).include?(5)
  puts "5 not in exclusive range"
end

# Works with variables
x = (1..10).include?(5)
puts x

# Plain parenthesized condition still works
y = 10
if (y > 5)
  puts "greater"
end

#@ expect:
# includes 5
# 5 not in exclusive range
# true
# greater
