#@ run-pass
#@ check-output
#
# Test Array mutation methods: push, pop, shift, unshift

def main
  arr = [2, 3, 4]

  # unshift adds to the beginning
  arr.unshift(1)
  puts "after unshift: #{arr.length}"

  # push adds to the end
  arr.push(5)
  puts "after push: #{arr.length}"

  # shift removes from the beginning
  first = arr.shift
  puts "shifted: #{first}"
  puts "after shift: #{arr.length}"

  # pop removes from the end
  last = arr.pop
  puts "popped: #{last}"
  puts "after pop: #{arr.length}"

  # print remaining elements
  arr.each do |n|
    puts n
  end
end

#@ expect:
# after unshift: 4
# after push: 5
# shifted: 1
# after shift: 4
# popped: 5
# after pop: 3
# 2
# 3
# 4
