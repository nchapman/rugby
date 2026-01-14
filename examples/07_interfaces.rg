# Rugby Interfaces
# Demonstrates: interface, implements, structural typing, is_a?, as

# Define interfaces
interface Speaker
  def speak -> String
end

interface Mover
  def move
end

# Interface composition
interface Actor < Speaker, Mover
end

# Explicit interface implementation (optional, but enables compile-time check)
class Dog implements Speaker
  def initialize(@name : String)
  end

  def speak -> String
    "Woof! I'm #{@name}"
  end

  def move
    puts "#{@name} runs on four legs"
  end
end

# Structural typing - satisfies Speaker without explicit implements
class Robot
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

  # Both satisfy Speaker - structural typing
  greet(dog)
  greet(robot)

  # Type checking with is_a?
  if dog.is_a?(Speaker)
    puts "Dog is a Speaker"
  end

  # Type assertion with as (returns optional)
  if let speaker = dog.as(Speaker)
    puts "Cast succeeded: #{speaker.speak}"
  end

  # Movement
  dog.move
  robot.move
end
