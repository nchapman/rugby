#@ run-pass
#@ check-output
#
# Test: Section 5.8 - Map with implicit value shorthand

name = "Alice"
age = 30

# Implicit value: {name:} same as {name: name}
m = {name:, age:}
puts m["name"]
puts m["age"]

#@ expect:
# Alice
# 30
