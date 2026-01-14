#@ run-pass
#@ check-output
#
# Test method chaining with self return

class Builder
  getter result : String

  def initialize
    @result = ""
  end

  def add(s : String) -> Builder
    @result = @result + s
    self
  end

  def space -> Builder
    @result = @result + " "
    self
  end
end

b = Builder.new
b.add("hello").space.add("world")
puts b.result

# Chain all at once
b2 = Builder.new.add("a").add("b").add("c")
puts b2.result

#@ expect:
# hello world
# abc
