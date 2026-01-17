#@ compile-fail
# Bug: Array indexing with variable index returns 'any' type
# When using a variable index (not literal), typed Array<T> returns 'any'
# This blocks 3/5 benchmarks

flags : Array<Bool> = [true, false, true]
i = 0

# Expected behavior: flags[i] should be Bool
# Actual: flags[i] is 'any', so ! operator fails

if !flags[i]
  puts "first is false"
end
#~ ERROR: operator ! not defined on.*any
