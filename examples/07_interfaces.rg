# Rugby Interfaces
# Demonstrates: interface, implements, is_a?, as (spec 7.6, 9)

# Define interfaces (spec 9)
interface Speaker
  def speak -> String
end

interface Mover
  def move
end

# Interface composition (spec 9)
interface Actor < Speaker, Mover
end

# Class explicitly implementing an interface (spec 7.6)
class Dog implements Speaker
  @name : String

  def initialize(@name : String)
  end

  def speak -> String
    "Woof! I'm #{@name}"
  end

  def move
    puts "#{@name} runs on four legs"
  end
end

# Class satisfying interface structurally (spec 7.6)
class Robot
  @model : String

  def initialize(@model : String)
  end

  def speak -> String
    "Beep boop, I am #{@model}"
  end

  def move
    puts "#{@model} rolls on wheels"
  end
end

# Function accepting interface type
def greet(s : Speaker)
  puts s.speak
end

def main
  dog = Dog.new("Rex")
  robot = Robot.new("R2D2")

  # Direct method calls
  puts dog.speak
  puts robot.speak

  # Pass to function expecting interface (structural typing)
  greet(dog)
  greet(robot)

  # Type checking with is_a? (spec 9.2)
  if dog.is_a?(Speaker)
    puts "Dog is a Speaker"
  end

  # Type assertion with as (spec 9.2)
  if let speaker = dog.as(Speaker)
    puts "Cast successful: #{speaker.speak}"
  end

  # Movement
  dog.move
  robot.move
end
