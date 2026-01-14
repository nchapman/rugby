#@ run-pass
#@ check-output
#
# Test: If/elsif/else control flow

x = 10

if x > 15
  puts "big"
elsif x > 5
  puts "medium"
else
  puts "small"
end

# Unless
unless x < 5
  puts "x is not small"
end

#@ expect:
# medium
# x is not small
