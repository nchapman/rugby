#@ run-pass
#@ check-output
#
# Test: Extended String methods coverage

# chars
chars = "abc".chars
puts chars.length
puts chars[0]
puts chars[1]
puts chars[2]

# reverse
puts "hello".reverse

# lstrip and rstrip
puts "  left".lstrip
puts "right  ".rstrip

# capitalize
puts "hELLO".capitalize

# replace (gsub)
puts "hello world".replace("world", "ruby")

# lines
text = "line1\nline2\nline3"
lines = text.lines
puts lines.length
puts lines[0]

# length
puts "hello".length

# Negative indexing
s = "hello"
puts s[-1]
puts s[-2]

#@ expect:
# 3
# a
# b
# c
# olleh
# left
# right
# Hello
# hello ruby
# 3
# line1
# 5
# o
# l
