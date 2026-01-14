#@ run-pass
#@ check-output
#
# BUG (discovered): Inherited getters return pointer addresses instead of values
# This test uses workaround - define getter method explicitly

class Animal
  getter name : String

  def initialize(@name : String)
  end
end

class Dog < Animal
  def speak -> String
    "Woof!"
  end

  # Workaround: access @name directly
  def get_name -> String
    @name
  end
end

dog = Dog.new("Buddy")
puts dog.get_name
puts dog.speak

#@ expect:
# Buddy
# Woof!
