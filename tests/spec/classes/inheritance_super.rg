#@ run-pass
#@ check-output
#
# Test: Section 11.6 - Inheritance and super

class Animal
  getter name: String

  def initialize(@name: String)
  end

  def speak: String
    "..."
  end

  def describe: String
    "Animal: #{@name}"
  end
end

class Dog < Animal
  getter breed: String

  def initialize(name: String, @breed: String)
    super(name)
  end

  def speak: String
    "Woof!"
  end

  def describe: String
    "#{super} (#{@breed})"
  end
end

dog = Dog.new("Rex", "Labrador")
puts dog.name
puts dog.breed
puts dog.speak
puts dog.describe

#@ expect:
# Rex
# Labrador
# Woof!
# Animal: Rex (Labrador)
