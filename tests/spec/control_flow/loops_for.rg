#@ run-pass
#@ check-output
#
# Test: Section 9.7 - For loops

# Iterate collection
items = ["a", "b", "c"]
for item in items
  puts item
end

# Iterate range
for i in 1..3
  puts i
end

# Iterate map
m = {"x" => 1, "y" => 2}
for key, value in m
  puts "#{key}=#{value}"
end

#@ expect:
# a
# b
# c
# 1
# 2
# 3
# x=1
# y=2
