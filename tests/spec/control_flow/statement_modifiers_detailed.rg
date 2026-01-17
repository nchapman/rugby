#@ run-pass
#@ check-output
#
# Test: Section 9.8 - Statement modifiers

# return if
def check_id(id: Int): String
  return "invalid" if id < 0
  return "zero" if id == 0
  "valid"
end

puts check_id(-1)
puts check_id(0)
puts check_id(5)

# puts unless
valid = true
puts "error" unless valid

invalid = false
puts "error" unless invalid

# break/next with modifiers
for i in 1..5
  next if i == 2
  break if i == 4
  puts i
end

#@ expect:
# invalid
# zero
# valid
# error
# 1
# 3
