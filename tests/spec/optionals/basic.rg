#@ run-pass
#@ check-output
#
# Test: Basic optional type usage

def find_user(id: Int): String?
  return "Alice" if id == 1
  nil
end

# Found user
result = find_user(1)
if result.present?
  puts "Found"
end

# Not found
result2 = find_user(99)
if result2.nil?
  puts "Not found"
end

#@ expect:
# Found
# Not found
