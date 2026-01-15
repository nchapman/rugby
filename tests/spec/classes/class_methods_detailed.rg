#@ run-pass
#@ check-output
#
# Test: Section 11.5 - Class methods and class variables

class Counter
  @@count = 0

  def initialize
    @@count += 1
  end

  def self.count -> Int
    @@count
  end

  def self.reset
    @@count = 0
  end
end

puts Counter.count

c1 = Counter.new
puts Counter.count

c2 = Counter.new
puts Counter.count

Counter.reset
puts Counter.count

#@ expect:
# 0
# 1
# 2
# 0
