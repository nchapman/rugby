#@ compile-fail
#
# BUG-046: Map literal in method body
# Map literals with symbol shorthand in method bodies fail to parse

class Todo
  getter id : Int
  getter title : String

  def initialize(@id : Int, @title : String)
  end

  def to_map -> Map[String, any]
    {
      id: @id,      #~ expected
      title: @title
    }
  end
end
