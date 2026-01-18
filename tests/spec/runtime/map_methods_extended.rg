#@ run-pass
#@ check-output
#
# Test: Extended Map/Hash methods coverage

# Basic map operations
m = {"a" => 1, "b" => 2, "c" => 3}

# keys and values
keys = m.keys.sorted
puts keys.length
puts keys[0]

values = m.values.sorted
puts values.length
puts values[0]

# has_key?
puts m.has_key?("a")
puts m.has_key?("z")

# empty?
puts {}.empty?
puts m.empty?

# size/length
puts m.size
puts m.length

# fetch with default
puts m.fetch("a", 99)
puts m.fetch("z", 99)

# merge
m2 = {"c" => 30, "d" => 4}
merged = m.merge(m2)
puts merged["c"]
puts merged["d"]
puts m["c"]  # Original unchanged

# select
evens = {"a" => 1, "b" => 2, "c" => 3, "d" => 4}.select do |k, v|
  v % 2 == 0
end
puts evens.size

# reject
odds = {"a" => 1, "b" => 2, "c" => 3, "d" => 4}.reject do |k, v|
  v % 2 == 0
end
puts odds.size

# invert
inverted = {"a" => 1, "b" => 2}.invert
puts inverted[1]
puts inverted[2]

# each iteration
sum = 0
{"x" => 10, "y" => 20}.each do |k, v|
  sum += v
end
puts sum

#@ expect:
# 3
# a
# 3
# 1
# true
# false
# true
# false
# 3
# 3
# 1
# 99
# 30
# 4
# 3
# 2
# 2
# a
# b
# 30
