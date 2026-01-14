#@ run-pass
#@ check-output
#
# Test visibility modifiers (pub)

pub class Counter
  getter value : Int

  def initialize
    @value = 0
  end

  # pub makes method exportable for multi-file projects
  pub def increment
    @value = @value + 1
  end

  pub def decrement
    @value = @value - 1
  end
end

c = Counter.new
c.increment
c.increment
c.increment
c.decrement
puts c.value

#@ expect:
# 2
