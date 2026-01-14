#@ run-pass
#@ check-output
#
# Test case_type statements for type matching
# TODO: case_type should implicitly return values (Ruby expression semantics)
#       Currently requires explicit `return` in each branch

def describe_value(x : any) -> String
  case_type x
  when String
    return "a string: #{x}"
  when Int
    return "an integer: #{x}"
  when Bool
    if x
      return "true boolean"
    else
      return "false boolean"
    end
  else
    return "unknown type"
  end
end

# Test with different types
puts describe_value("hello")
puts describe_value(42)
puts describe_value(true)
puts describe_value(false)

#@ expect:
# a string: hello
# an integer: 42
# true boolean
# false boolean
