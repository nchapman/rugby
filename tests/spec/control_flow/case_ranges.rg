#@ run-pass
#@ check-output
#
# Test: Section 9.5 - Case with ranges

status = 404

case status
when 200, 201
  puts "ok"
when 400..499
  puts "client error"
when 500..599
  puts "server error"
else
  puts "unknown"
end

# Another example
score = 75
case score
when 90..100
  puts "A"
when 80..89
  puts "B"
when 70..79
  puts "C"
else
  puts "F"
end

#@ expect:
# client error
# C
