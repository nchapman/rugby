#@ run-pass
#@ check-output
#
# Test: Section 4.1 - Comments
# Single line comments using #

# This is a comment
puts "before" # inline comment
puts "after"

# Multiple comment lines
# in a row
# are fine
puts "done"

#@ expect:
# before
# after
# done
