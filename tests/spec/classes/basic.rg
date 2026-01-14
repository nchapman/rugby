#@ run-pass
#@ check-output
#
# Test: Basic class definitions

class Person
  getter name : String
  getter age : Int

  def initialize(@name : String, @age : Int)
  end

  def greet -> String
    "Hi, I'm #{@name}"
  end
end

person = Person.new("Alice", 30)
puts person.name
puts person.age
puts person.greet

#@ expect:
# Alice
# 30
# Hi, I'm Alice
