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

# Iterate map (order is non-deterministic in Go)
m = {"a" => 1}
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
# a=1
