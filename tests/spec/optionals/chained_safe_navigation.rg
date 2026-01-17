#@ run-pass
#@ check-output
# Chained safe navigation works on nested optional fields

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

# Test 1: Full chain with all values present
city1 = City.new("Boston")
addr1 = Address.new(city1)
user1 = User.new(addr1)
name1 = user1.address&.city&.name ?? "Unknown"
puts name1

# Test 2: Address is nil
user2 = User.new(nil)
name2 = user2.address&.city&.name ?? "NoAddress"
puts name2

# Test 3: City is nil
addr3 = Address.new(nil)
user3 = User.new(addr3)
name3 = user3.address&.city&.name ?? "NoCity"
puts name3

# Test 4: Different city
city4 = City.new("Seattle")
addr4 = Address.new(city4)
user4 = User.new(addr4)
name4 = user4.address&.city&.name ?? "Unknown"
puts name4

#@ expect:
# Boston
# NoAddress
# NoCity
# Seattle
