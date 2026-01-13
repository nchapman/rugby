# Rugby Classes
# Demonstrates: class, initialize, fields, accessors, inheritance, pub (spec 7, 10)

class Point
  # Explicit field declarations (spec 7)
  @x : Int
  @y : Int

  # Initialize with parameter promotion (spec 7)
  def initialize(@x : Int, @y : Int)
  end

  # Instance methods access fields with @
  def magnitude_squared -> Int
    @x * @x + @y * @y
  end

  def to_s -> String
    "(#{@x}, #{@y})"
  end

  # Explicit self usage (spec 7)
  def equals(other : Point) -> Bool
    self.magnitude_squared == other.magnitude_squared
  end
end

class Person
  # Using accessor declarations (spec 7.4)
  getter name : String
  getter age : Int
  property email : String

  def initialize(@name : String, @age : Int, @email : String)
  end

  def introduce -> String
    "Hi, I'm #{@name}, #{@age} years old"
  end

  def have_birthday
    @age = @age + 1
  end
end

# Inheritance with < (spec 7.5)
class Animal
  @name : String

  def initialize(@name : String)
  end

  def speak -> String
    "..."
  end

  def name -> String
    @name
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

# pub exports to Go (spec 10)
pub class Counter
  @count : Int

  def initialize
    @count = 0
  end

  pub def inc
    @count += 1
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
  puts "p1.equals(p2): #{p1.equals(p2)}"

  # Create Person with accessors
  person = Person.new("Alice", 30, "alice@example.com")
  puts person.introduce

  # Use getter
  puts "Name: #{person.name}"
  puts "Age: #{person.age}"

  # Use property getter and setter
  puts "Email: #{person.email}"
  person.email = "alice.new@example.com"
  puts "New email: #{person.email}"

  # Modify through method
  person.have_birthday
  puts "After birthday: #{person.age}"

  # Inheritance (spec 7.5)
  cat = Cat.new("Whiskers")
  dog = Dog.new("Rex")
  puts cat.speak
  puts dog.speak

  # Exported class
  counter = Counter.new
  counter.inc
  counter.inc
  puts "Counter: #{counter.value}"
end
