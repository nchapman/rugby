#@ run-pass
#@ check-output
#
# Test: Section 13.2 - Structural typing

interface Speaker
  def speak -> String
end

# No "implements" needed - just have matching methods
class Dog
  def speak -> String
    "Woof!"
  end
end

class Cat
  def speak -> String
    "Meow!"
  end
end

class Robot
  def speak -> String
    "Beep boop"
  end
end

def announce(s : Speaker)
  puts s.speak
end

# All satisfy Speaker structurally
announce(Dog.new)
announce(Cat.new)
announce(Robot.new)

#@ expect:
# Woof!
# Meow!
# Beep boop
