#@ run-pass
#@ check-output
#
# Test: Section 9.6 - case_type for type matching

# Test with binding variables
def print_type(obj : Any)
  case_type obj
  when s : String
    puts "string: #{s}"
  when n : Int
    puts "int: #{n}"
  when b : Bool
    puts "bool: #{b}"
  else
    puts "unknown type"
  end
end

print_type("hello")
print_type(42)
print_type(true)
print_type(3.14)

# Test without binding variables
def type_name(obj : Any)
  case_type obj
  when String
    puts "is string"
  when Int
    puts "is int"
  else
    puts "is other"
  end
end

type_name("test")
type_name(100)
type_name([1, 2, 3])

#@ expect:
# string: hello
# int: 42
# bool: true
# unknown type
# is string
# is int
# is other
