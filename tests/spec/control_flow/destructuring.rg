#@ run-pass
#@ check-output
#
# Test: Section 9.3 - Destructuring assignment

# Tuple destructuring
def get_pair: (Int, String)
  42, "hello"
end

a, b = get_pair
puts a
puts b

# Splat in destructuring
items = [1, 2, 3, 4, 5]
first, *rest = items
puts first
puts rest.length

# Map destructuring
user_data = {name: "Alice", age: 30}
{name:, age:} = user_data
puts name
puts age

#@ expect:
# 42
# hello
# 1
# 4
# Alice
# 30
