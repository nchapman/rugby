#@ run-pass
#@ check-output
#
# Test: Section 18.5 - Defer

def example
  defer puts("first")
  defer puts("second")
  defer puts("third")
  puts "body"
end

example

#@ expect:
# body
# third
# second
# first
