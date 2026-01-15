#@ run-pass
#@ check-output
#
# Test: Section 14.4 - Module conflicts (diamond problem)

module A
  def greet -> String
    "Hello from A"
  end
end

module B
  def greet -> String
    "Hello from B"
  end
end

class C
  include A
  include B  # B's greet wins (last included)
end

class D
  include A
  include B

  def greet -> String
    "Hello from D"  # class method wins
  end
end

puts C.new.greet
puts D.new.greet

#@ expect:
# Hello from B
# Hello from D
