#@ run-pass
#@ check-output
#
# Test: Section 5.8 - Map literals with symbol shorthand

class Todo
  getter id : Int
  getter title : String

  def initialize(@id : Int, @title : String)
  end

  def to_map -> Map<String, Any>
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
