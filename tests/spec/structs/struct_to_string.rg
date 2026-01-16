#@ run-pass
#@ check-output
#
# Test: Section 12.2 - Struct string representation
# TODO: Implement auto-generated String() method for structs

struct User
  id : Int64
  name : String
  active : Bool
end

user = User{id: 1, name: "Alice", active: true}

# Structs have auto-generated string representation
puts user

# String interpolation uses the string representation
message = "User: #{user}"
puts message

#@ expect:
# User{id: 1, name: "Alice", active: true}
# User: User{id: 1, name: "Alice", active: true}
