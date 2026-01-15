#@ run-pass
#@ check-output
#
# Test times, upto, downto integer iteration methods

# Test times - iterate n times (0 to n-1)
3.times -> (i) { puts "times: #{i}" }

# Test upto - iterate from start to end (inclusive)
1.upto(3) -> (i) { puts "upto: #{i}" }

# Test downto - iterate from start down to end (inclusive)
3.downto(1) -> (i) { puts "downto: #{i}" }

# Edge cases: zero iterations
0.times -> (i) { puts "should not print" }
5.upto(3) -> (i) { puts "should not print" }
1.downto(5) -> (i) { puts "should not print" }
puts "edge cases passed"

#@ expect:
# times: 0
# times: 1
# times: 2
# upto: 1
# upto: 2
# upto: 3
# downto: 3
# downto: 2
# downto: 1
# edge cases passed
