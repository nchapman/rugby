# Rugby Classes
# Demonstrates: class, initialize, fields, accessors, inheritance, self, pub

# Simple class with parameter promotion
class Point
  def initialize(@x : Int, @y : Int)
  end

  def magnitude_squared -> Int
    @x * @x + @y * @y
  end

  def to_s -> String
    "(#{@x}, #{@y})"
  end

  # Use self for method chaining or explicit reference
  def equals?(other : Point) -> Bool
    self.magnitude_squared == other.magnitude_squared
  end
end

# Class with accessors
class Person
  getter name : String      # def name -> String
  getter age : Int          # def age -> Int
  property email : String   # getter + setter

  def initialize(@name : String, @age : Int, @email : String)
  end

  def introduce -> String
    "Hi, I'm #{@name}, #{@age} years old"
  end

  def have_birthday
    @age += 1
  end
end

# Inheritance with <
class Animal
  getter name : String

  def initialize(@name : String)
  end

  def speak -> String
    "..."
  end
end

class Cat < Animal
  def speak -> String
    "Meow! I'm #{@name}"
  end
end

class Dog < Animal
  def speak -> String
    "Woof! I'm #{@name}"
  end
end

# pub exports to Go (uppercase in generated code)
pub class Counter
  @count : Int

  def initialize
    @count = 0
  end

  pub def inc
    @count += 1
    self  # return self for chaining
  end

  pub def value -> Int
    @count
  end
end

def main
  # Create Point instances
  p1 = Point.new(3, 4)
  p2 = Point.new(4, 3)
  puts "Point: #{p1.to_s}"
  puts "Magnitude squared: #{p1.magnitude_squared}"
  puts "p1.equals?(p2): #{p1.equals?(p2)}"

  # Person with accessors
  person = Person.new("Alice", 30, "alice@example.com")
  puts person.introduce

  puts "Name: #{person.name}"
  puts "Email: #{person.email}"

  person.email = "alice.new@example.com"
  puts "New email: #{person.email}"

  person.have_birthday
  puts "After birthday: #{person.age}"

  # Inheritance and polymorphism
  cat = Cat.new("Whiskers")
  dog = Dog.new("Rex")
  puts cat.speak
  puts dog.speak

  # Method chaining with self
  counter = Counter.new
  counter.inc.inc.inc
  puts "Counter: #{counter.value}"
end
