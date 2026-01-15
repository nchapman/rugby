#@ run-pass
#@ check-output
#
# Test: Section 19.6 - Global functions (Kernel)

# puts - print with newline
puts "hello"
puts 42
puts true

# print - print without newline (can't easily test)

# p - debug print
p "debug"
p [1, 2, 3]

#@ expect:
# hello
# 42
# true
# "debug"
# [1 2 3]
