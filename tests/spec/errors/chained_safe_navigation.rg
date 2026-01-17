#@ compile-fail
# Bug: Chained safe navigation on nested optional class fields
# Using &. multiple times on nested optional fields causes codegen errors

class City
  getter name : String
  def initialize(@name: String)
  end
end

class Address
  getter city : City?
  def initialize(@city: City?)
  end
end

class User
  getter address : Address?
  def initialize(@address: Address?)
  end
end

# Create user with full chain
city = City.new("Boston")
addr = Address.new(city)
user = User.new(addr)

# Expected behavior: should work
# name = user.address&.city&.name ?? "Unknown"

# Current: codegen error "cannot indirect variable of interface type any"
name = user.address&.city&.name ?? "Unknown"
puts name
#~ ERROR: cannot indirect
