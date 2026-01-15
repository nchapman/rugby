#@ run-pass
#@ check-output
#
# Test: Section 8.4 - The nil keyword

# nil valid for optional types
def find_user(id : Int) -> String?
  return nil if id < 0
  "User#{id}"
end

puts find_user(-1) ?? "none"
puts find_user(1) ?? "none"

# nil comparison
result = find_user(-1)
if result == nil
  puts "is nil"
end

result2 = find_user(1)
if result2 != nil
  puts "not nil"
end

#@ expect:
# none
# User1
# is nil
# not nil
