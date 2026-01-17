#@ run-pass
#@ check-output
#
# Test using 'next' keyword as a method name (common in iterators/generators)

class Random
  property seed: Int

  def initialize(seed: Int)
    @seed = seed
  end

  # 'next' is a Rugby keyword but NOT a Go keyword, so it can be used as a method name
  def next: Int
    @seed = (@seed * 1103515245 + 12345) % 2147483648
    @seed
  end
end

class Iterator
  property value: Int

  def initialize(value: Int)
    @value = value
  end

  def next: Int
    @value += 1
    @value
  end
end

def main
  rng = Random.new(42)
  puts rng.next
  puts rng.next

  iter = Iterator.new(10)
  puts iter.next
  puts iter.next
end

#@ expect:
# 1250496027
# 1116302264
# 11
# 12
