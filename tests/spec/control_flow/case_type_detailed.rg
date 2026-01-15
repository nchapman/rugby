#@ run-pass
#@ check-output
#
# Test: Section 9.6 - case_type for type matching

interface Printable
  def to_string -> String
end

class Dog
  def to_string -> String
    "Dog"
  end
end

class Cat
  def to_string -> String
    "Cat"
  end
end

def describe(obj : Printable) -> String
  case_type obj
  when d : Dog
    return "It's a dog: #{d.to_string}"
  when c : Cat
    return "It's a cat: #{c.to_string}"
  else
    return "Unknown animal"
  end
end

puts describe(Dog.new)
puts describe(Cat.new)

#@ expect:
# It's a dog: Dog
# It's a cat: Cat
