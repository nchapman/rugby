#@ run-pass
#@ check-output
#
# Test word array literals (%w)

# Basic word array
words = %w[foo bar baz]
words.each { |w| puts w }

# Word array with different delimiters
colors = %w(red green blue)
colors.each { |c| puts c }

# Using word arrays
fruits = %w[apple banana cherry]
puts fruits.length
puts fruits[1]

#@ expect:
# foo
# bar
# baz
# red
# green
# blue
# 3
# banana
