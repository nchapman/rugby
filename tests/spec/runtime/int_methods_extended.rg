#@ run-pass
#@ check-output
#
# Test: Extended Integer methods coverage

# abs
puts (-5).abs
puts 5.abs

# clamp
puts 15.clamp(0, 10)
puts (-5).clamp(0, 10)
puts 5.clamp(0, 10)

# predicates
puts 4.even?
puts 5.even?
puts 4.odd?
puts 5.odd?

puts 5.positive?
puts 0.positive?
puts (-5).positive?

puts (-5).negative?
puts 0.negative?
puts 5.negative?

puts 0.zero?
puts 5.zero?

# times
sum = 0
3.times do |i|
  sum += i
end
puts sum

# upto
total = 0
1.upto(5) do |n|
  total += n
end
puts total

# downto
countdown = 0
5.downto(1) do |n|
  countdown += n
end
puts countdown

# to_f
puts 42.to_f

# to_s
puts 123.to_s

#@ expect:
# 5
# 5
# 10
# 0
# 5
# true
# false
# false
# true
# true
# false
# false
# true
# false
# false
# true
# false
# 3
# 15
# 15
# 42
# 123
