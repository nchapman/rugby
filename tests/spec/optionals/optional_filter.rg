#@ run-pass
#@ check-output

# Test optional.filter method
# filter -> { |x| pred(x) } returns Some(x) if present and predicate is true, else None

# Test 1: Filter with passing predicate on Some value
name : String? = "Alice"
result1 = name.filter -> { |n| n.length > 3 }
if let r = result1
  puts "1: #{r}"
else
  puts "1: none"
end

# Test 2: Filter with failing predicate on Some value
name2 : String? = "Al"
result2 = name2.filter -> { |n| n.length > 3 }
if let r = result2
  puts "2: #{r}"
else
  puts "2: none"
end

# Test 3: Filter on None value
name3 : String? = nil
result3 = name3.filter -> { |n| n.length > 0 }
if let r = result3
  puts "3: #{r}"
else
  puts "3: none"
end

# Test 4: Filter with explicit type annotation
typed : String? = "Hello"
filtered : String? = typed.filter -> { |s| s.length > 4 }
if let f = filtered
  puts "4: #{f}"
else
  puts "4: none"
end

#@ expect:
# 1: Alice
# 2: none
# 3: none
# 4: Hello
