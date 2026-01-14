#@ run-pass
#@ check-output
#
# Test getter, setter, and property declarations

class Person
  getter name : String
  setter age : Int
  property email : String

  def initialize(@name : String, @age : Int, @email : String)
  end

  def get_age -> Int
    @age
  end
end

person = Person.new("Alice", 30, "alice@example.com")

# Test getter (read-only)
puts "name: #{person.name}"

# Test setter (write-only, need internal method to read)
person.age = 31
puts "age: #{person.get_age}"

# Test property (read and write)
puts "email: #{person.email}"
person.email = "alice@newmail.com"
puts "new email: #{person.email}"

#@ expect:
# name: Alice
# age: 31
# email: alice@example.com
# new email: alice@newmail.com
