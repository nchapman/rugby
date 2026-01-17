#@ run-pass
#@ check-output
#
# Test: Type assertions with class types (not just interfaces)

class Dog
  def speak: String
    "Woof"
  end
end

class Cat
  def speak: String
    "Meow"
  end
end

def test_dog(obj: Any)
  # is_a? with class type
  if obj.is_a?(Dog)
    puts "is a dog"
  end

  # as with class type
  if let dog = obj.as(Dog)
    puts dog.speak
  end
end

def test_cat(obj: Any)
  if obj.is_a?(Cat)
    puts "is a cat"
  end

  if let cat = obj.as(Cat)
    puts cat.speak
  end
end

test_dog(Dog.new)
test_cat(Cat.new)

#@ expect:
# is a dog
# Woof
# is a cat
# Meow
