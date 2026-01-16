#@ run-pass
#@ check-output
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

# from_string returns Color?
color = Color.from_string("Red")
if let c = color
  puts c.to_s
end

#@ expect:
# Red
# 3
# Red
