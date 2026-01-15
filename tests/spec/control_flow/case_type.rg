#@ run-pass
#@ check-output
#
# Test: Section 9.6 - case_type for type matching
# TODO: case_type should implicitly return values (Ruby expression semantics)
#       Currently requires explicit `return` in each branch

def describe_value(x : Any) -> String
  case_type x
  when s : String
    return "a string: #{s}"
  when n : Int
    return "an integer: #{n}"
  when b : Bool
    if b
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
