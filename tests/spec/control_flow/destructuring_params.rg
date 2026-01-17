#@ run-pass
#@ check-output
#
# Test: Section 9.3 - Destructuring in function parameters

struct UserData
  name: String
  age: Int
end

# Struct destructuring in function parameters
def process({name:, age:} : UserData)
  puts "#{name} is #{age}"
end

data = UserData{name: "Alice", age: 30}
process(data)

# Destructuring with rename
def process_renamed({name: n, age: a} : UserData)
  puts "#{n} (age #{a})"
end

process_renamed(data)

#@ expect:
# Alice is 30
# Alice (age 30)
