#@ run-pass
#@ check-output
#
# Test: Class inheritance with methods

class Animal
  def initialize
  end

  def speak -> String
    "..."
  end
end

class Dog < Animal
  def speak -> String
    "Woof!"
  end
end

class Cat < Animal
  def speak -> String
    "Meow!"
  end
end

dog = Dog.new
cat = Cat.new

puts dog.speak
puts cat.speak

#@ expect:
# Woof!
# Meow!
