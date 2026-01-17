#@ run-pass
#@ check-output
#
# Test: Section 8.3 - if let scoping rules

class User
  getter name: String
  def initialize(@name: String)
  end
end

def find_user(id: Int): User?
  if id == 1
    return User.new("Alice")
  end
  nil
end

# Outer variable
user = User.new("Outer")

# if let creates new binding, shadows outer
if let user = find_user(1)
  puts user.name
end

# Outer user unchanged after block
puts user.name

# if let with else
if let u = find_user(99)
  puts u.name
else
  puts "not found"
end

#@ expect:
# Alice
# Outer
# not found
