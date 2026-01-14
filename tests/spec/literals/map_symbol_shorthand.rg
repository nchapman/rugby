#@ run-pass
#@ check-output
#
# Test that map literals with symbol shorthand work in method bodies

class Todo
  getter id : Int
  getter title : String

  def initialize(@id : Int, @title : String)
  end

  def to_map -> Map[String, any]
    {
      id: @id,
      title: @title
    }
  end
end

todo = Todo.new(1, "Buy milk")
m = todo.to_map
puts m["id"]
puts m["title"]

#@ expect:
# 1
# Buy milk
