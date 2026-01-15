#@ run-pass
#@ check-output
#
# Test: Section 9.9 - Loop modifiers

# Statement while
arr = [1, 2, 3]
puts arr.shift while arr.length > 0
puts "done"

# Statement until
arr2 = [4, 5]
puts arr2.shift until arr2.empty?
puts "done2"

#@ expect:
# 1
# 2
# 3
# done
# 4
# 5
# done2
