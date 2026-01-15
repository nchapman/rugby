#@ run-pass
#@ check-output
#
# Test find, any?, all?, none? with lambda

numbers = [1, 2, 3, 4, 5, 6]

# Test find - returns first matching element
found = numbers.find -> (n) { n > 3 }
if let f = found
  puts "found: #{f}"
end

found2 = numbers.find -> (n) { n > 100 }
if let f = found2
  puts "found big"
else
  puts "not found"
end

# Test any? - returns true if any element matches
puts "any even? #{numbers.any? -> (n) { n.even? }}"
puts "any > 10? #{numbers.any? -> (n) { n > 10 }}"

# Test all? - returns true if all elements match
puts "all positive? #{numbers.all? -> (n) { n > 0 }}"
puts "all even? #{numbers.all? -> (n) { n.even? }}"

# Test none? - returns true if no elements match
puts "none negative? #{numbers.none? -> (n) { n < 0 }}"
puts "none even? #{numbers.none? -> (n) { n.even? }}"

#@ expect:
# found: 4
# not found
# any even? true
# any > 10? false
# all positive? true
# all even? false
# none negative? true
# none even? false
