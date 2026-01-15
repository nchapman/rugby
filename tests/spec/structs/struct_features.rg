#@ compile-fail
#@ skip: Structs not yet implemented (Section 12.2)
#
# Test: Section 12.2 - Struct features (auto-constructor, equality)
# TODO: Implement struct features

struct User
  id : Int64
  name : String
  email : String
end

u1 = User{id: 1, name: "Alice", email: "alice@example.com"}
u2 = User{id: 1, name: "Alice", email: "alice@example.com"}

# Value equality
puts u1 == u2

# Can be map keys
users = Map<User, Int>{}
users[u1] = 1
