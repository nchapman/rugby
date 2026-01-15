#@ run-pass
#@ check-output
#
# Test set literals and operations (Section 5.9)

def main
  # Basic set literal
  s1 = Set{1, 2, 3}

  # Empty set
  empty = Set<String>{}

  # Test include?
  puts s1.include?(1)
  puts s1.include?(10)

  # Test size
  puts s1.size

  # Test add
  s1.add(100)
  puts s1.include?(100)

  # Test delete
  s1.delete(1)
  puts s1.include?(1)

  # Test empty?
  puts empty.empty?
  puts s1.empty?

  # Test any?
  puts s1.any?
  puts empty.any?

  # Test union (|)
  union = Set{1, 2} | Set{2, 3}
  puts union.size

  # Test intersection (&)
  intersect = Set{1, 2, 3} & Set{2, 3, 4}
  puts intersect.size

  # Test difference (-)
  diff = Set{1, 2, 3} - Set{2}
  puts diff.size

  # Test each iteration
  sum = 0
  Set{10, 20, 30}.each do |x|
    sum = sum + x
  end
  puts sum
end

#@ expect:
# true
# false
# 3
# true
# false
# true
# false
# true
# false
# 3
# 2
# 2
# 60
