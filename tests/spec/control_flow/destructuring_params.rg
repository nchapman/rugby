#@ run-pass
#@ check-output
#@ skip: Destructuring not yet implemented (Section 9.3)
#
# Test: Section 9.3 - Destructuring in function parameters
# TODO: Implement parameter destructuring

struct UserData
  name : String
  age : Int
end

# Map destructuring in function parameters
def process({name:, age:} : UserData)
  puts "#{name} is #{age}"
end

data = UserData{name: "Alice", age: 30}
process(data)

# Also works with regular maps
def greet({name:} : Map<Symbol, String>)
  puts "Hello, #{name}!"
end

info = {name: "Bob"}
greet(info)

#@ expect:
# Alice is 30
# Hello, Bob!
