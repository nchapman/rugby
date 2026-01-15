#@ run-pass
#@ check-output
#
# Test: Section 8.1 - Safe navigation operator (&.)

class Address
  getter city : String

  def initialize(@city : String)
  end
end

class User
  getter name : String
  getter address : Address?

  def initialize(@name : String, @address : Address?)
  end
end

# User with address
u1 = User.new("Alice", Address.new("NYC"))
puts u1.address&.city ?? "Unknown"

# User without address
u2 = User.new("Bob", nil)
puts u2.address&.city ?? "Unknown"

#@ expect:
# NYC
# Unknown
