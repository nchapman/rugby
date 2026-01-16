#@ run-pass
#@ check-output
#
# Test: Section 12.2 - Struct features (auto-constructor, equality)

struct User
  id : Int64
  name : String
  email : String
end

u1 = User{id: 1, name: "Alice", email: "alice@example.com"}
u2 = User{id: 1, name: "Alice", email: "alice@example.com"}
u3 = User{id: 2, name: "Bob", email: "bob@example.com"}

# Value equality - structs with same values are equal
puts u1 == u2
puts u1 == u3

#@ expect:
# true
# false
