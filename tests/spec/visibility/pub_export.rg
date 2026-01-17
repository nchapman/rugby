#@ run-pass
#@ check-output
#
# Test: Section 15.1 - Export with pub

# pub exports to Go (uppercase)
pub def double(x: Int): Int
  helper(x)
end

# No pub = internal (lowercase in Go)
def helper(x: Int): Int
  x * 2
end

pub class Counter
  pub getter count: Int

  def initialize
    @count = 0
  end

  pub def inc
    @count += 1
  end

  # internal method
  def reset
    @count = 0
  end
end

puts double(5)

c = Counter.new
c.inc
c.inc
puts c.count

#@ expect:
# 10
# 2
