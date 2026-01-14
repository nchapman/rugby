#@ run-pass
#@ check-output
#
# Test that getters defined in parent classes work correctly in child classes

class Animal
  getter name : String

  def initialize(@name : String)
  end
end

class Dog < Animal
  def speak -> String
    "Woof!"
  end
end

dog = Dog.new("Buddy")
puts dog.name
puts dog.speak

#@ expect:
# Buddy
# Woof!
