# Todo App - A complete idiomatic Rugby example
# Demonstrates: classes, blocks, optionals, error handling, file I/O
#
# This example shows what idiomatic Rugby looks like in practice.

import rugby/file
import rugby/json
import time

class Todo
  getter id : Int
  getter title : String
  getter created_at : String
  property completed : Bool

  def initialize(@id : Int, @title : String)
    @completed = false
    @created_at = time.now.format("2006-01-02 15:04:05")
  end

  def complete
    @completed = true
  end

  def to_map -> Map<String, any>
    {
      id: @id,
      title: @title,
      completed: @completed,
      created_at: @created_at
    }
  end
end

class TodoList
  def initialize
    @todos = [] : Array<Todo>
    @next_id = 1
  end

  def add(title : String) -> Todo
    todo = Todo.new(@next_id, title)
    @todos = @todos << todo
    @next_id += 1
    todo
  end

  def find(id : Int) -> Todo?
    @todos.find -> { |t| t.id == id }
  end

  def complete(id : Int) -> Bool
    if let todo = find(id)
      todo.complete
      true
    else
      false
    end
  end

  def remove(id : Int) -> Bool
    original_length = @todos.length
    @todos = @todos.reject -> { |t| t.id == id }
    @todos.length < original_length
  end

  def pending -> Array<Todo>
    @todos.reject(&:completed)
  end

  def completed -> Array<Todo>
    @todos.select(&:completed)
  end

  def all -> Array<Todo>
    @todos
  end

  def stats -> String
    total = @todos.length
    done = completed.length
    "#{done}/#{total} completed"
  end

  def save(path : String) -> Error
    data = @todos.map -> { |t| t.to_map }
    content = json.pretty(data)!
    file.write(path, content)
  end
end

def print_todos(todos : Array<Todo>, label : String)
  puts "\n#{label}:"
  return puts "  (none)" if todos.empty?

  todos.each -> do |todo|
    status = todo.completed ? "[x]" : "[ ]"
    puts "  #{status} ##{todo.id}: #{todo.title}"
  end
end

def main
  list = TodoList.new

  # Add some todos
  list.add("Learn Rugby syntax")
  list.add("Build something cool")
  list.add("Share with friends")

  print_todos(list.all, "All Todos")

  # Complete some tasks
  list.complete(1)
  list.complete(2)

  print_todos(list.pending, "Pending")
  print_todos(list.completed, "Completed")

  puts "\nStats: #{list.stats}"

  # Find a specific todo
  if let todo = list.find(3)
    puts "\nFound: #{todo.title}"
  end

  # Try to find non-existent todo
  missing = list.find(99)
  puts "Todo 99: #{missing.ok? ? "found" : "not found"}"

  # Remove a todo
  if list.remove(2)
    puts "\nRemoved todo #2"
  end

  print_todos(list.all, "After Removal")

  # Save to file (demonstrates error handling)
  err = list.save("/tmp/todos.json")
  if err == nil
    puts "\nSaved to /tmp/todos.json"
  else
    puts "\nFailed to save: #{err}"
  end
end
