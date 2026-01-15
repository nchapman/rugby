#@ run-pass
#@ check-output
#@ skip: Destructuring not yet implemented (Section 9.3)
#
# Test: Section 9.3 - Map destructuring
# TODO: Implement map destructuring

user_data = {name: "Alice", age: 30, city: "NYC"}

# Basic map destructuring (creates variables with same names)
{name:, age:} = user_data
puts name
puts age

# Map destructuring with rename
{name: user_name, age: user_age} = user_data
puts user_name
puts user_age

#@ expect:
# Alice
# 30
# Alice
# 30
