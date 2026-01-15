#@ run-pass
#@ check-output
#
# Test: Section 5.2/5.8 - Map<K, V> composite type

# Hash rocket syntax
m1 = {"name" => "Alice", "city" => "NYC"}
puts m1["name"]
puts m1["city"]

# Symbol shorthand syntax
m2 = {name: "Bob", age: 30}
puts m2["name"]
puts m2["age"]

# Map methods
puts m2.length
puts m2.has_key?("name")
puts m2.has_key?("missing")

# Empty typed map
empty : Map<String, Int> = {}
puts empty.length

#@ expect:
# Alice
# NYC
# Bob
# 30
# 2
# true
# false
# 0
