#@ run-pass
#@ check-output
#
# Test: If-let pattern for optional binding

def find_name(id : Int) -> String?
  return "Alice" if id == 1
  nil
end

if let name = find_name(1)
  puts "Hello, #{name}!"
else
  puts "No name"
end

if let name = find_name(99)
  puts "Hello, #{name}!"
else
  puts "No name"
end

#@ expect:
# Hello, Alice!
# No name
