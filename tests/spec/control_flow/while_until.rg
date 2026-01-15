#@ run-pass
#@ check-output
#
# Test while and until loops

# Test basic while loop
count = 0
while count < 3
  puts "while: #{count}"
  count += 1
end

# Test basic until loop
n = 3
until n == 0
  puts "until: #{n}"
  n -= 1
end

# Test while with complex condition
i = 0
arr : Array<Int> = []
while arr.length < 4
  arr = arr << i
  i += 1
end
puts "array length: #{arr.length}"

# Test until with method call condition
items = ["x", "y", "z"]
until items.empty?
  puts "item: #{items.shift}"
end

#@ expect:
# while: 0
# while: 1
# while: 2
# until: 3
# until: 2
# until: 1
# array length: 4
# item: x
# item: y
# item: z
