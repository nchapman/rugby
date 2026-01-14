#@ run-pass
#@ check-output
#
# Test break and next in loops

# Test break
for i in 0..10
  break if i == 3
  puts "break loop: #{i}"
end

# Test next
for i in 0..5
  next if i == 2
  puts "next loop: #{i}"
end

# Test break with unless
for i in 0..10
  break unless i < 4
  puts "break unless: #{i}"
end

# Test next with unless
for i in 0..3
  next unless i.even?
  puts "next unless: #{i}"
end

# Test break in while loop
i = 0
while i < 10
  break if i == 2
  puts "while break: #{i}"
  i += 1
end

# Test next in while loop
j = 0
while j < 5
  j += 1
  next if j == 3
  puts "while next: #{j}"
end

#@ expect:
# break loop: 0
# break loop: 1
# break loop: 2
# next loop: 0
# next loop: 1
# next loop: 3
# next loop: 4
# next loop: 5
# break unless: 0
# break unless: 1
# break unless: 2
# break unless: 3
# next unless: 0
# next unless: 2
# while break: 0
# while break: 1
# while next: 1
# while next: 2
# while next: 4
# while next: 5
