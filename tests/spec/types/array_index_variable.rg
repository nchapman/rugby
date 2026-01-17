#@ run-pass
#@ check-output
# Array indexing with variable index preserves element type
# When using a variable index, typed Array<T> returns T (not 'any')

flags : Array<Bool> = [true, false, true]
i = 0

# flags[i] should be Bool, so ! operator works
if !flags[i]
  puts "first is false"
else
  puts "first is true"
end

# Test with Int array
nums : Array<Int> = [10, 20, 30]
j = 1
puts nums[j] + 5

#@ expect:
# first is true
# 25
