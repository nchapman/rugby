#@ run-pass
#@ check-output
#
# Test continue keyword (alias for next)

# Test continue in for loop
for i in 0..5
  continue if i == 2
  puts "for loop: #{i}"
end

# Test continue in while loop
j = 0
while j < 5
  j += 1
  continue if j == 3
  puts "while loop: #{j}"
end

# Test continue with unless
for i in 0..3
  continue unless i.even?
  puts "unless: #{i}"
end

#@ expect:
# for loop: 0
# for loop: 1
# for loop: 3
# for loop: 4
# for loop: 5
# while loop: 1
# while loop: 2
# while loop: 4
# while loop: 5
# unless: 0
# unless: 2
