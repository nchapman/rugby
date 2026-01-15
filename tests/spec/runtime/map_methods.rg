#@ run-pass
#@ check-output
#
# Test: Section 19.2 - Map methods

m = {"a" => 1, "b" => 2, "c" => 3}

# keys / values
puts m.keys.length
puts m.values.length

# length / size
puts m.length

# empty?
puts m.empty?
puts {}.empty?

# has_key? / key?
puts m.has_key?("a")
puts m.has_key?("z")

# fetch with default
puts m.fetch("a", 0)
puts m.fetch("z", 99)

# get (returns optional)
puts m.get("a") ?? 0
puts m.get("z") ?? 0

#@ expect:
# 3
# 3
# 3
# false
# true
# true
# false
# 1
# 99
# 1
# 0
