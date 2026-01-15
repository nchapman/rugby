#@ run-pass
#@ check-output
#
# Test: Section 9.4 - Conditionals

# Basic if
if true
  puts "yes"
end

# if/else
x = 5
if x > 3
  puts "big"
else
  puts "small"
end

# if/elsif/else
score = 85
if score >= 90
  puts "A"
elsif score >= 80
  puts "B"
elsif score >= 70
  puts "C"
else
  puts "F"
end

# unless
valid = false
unless valid
  puts "invalid"
end

# Ternary
max = 5 > 3 ? 5 : 3
puts max

#@ expect:
# yes
# big
# B
# invalid
# 5
