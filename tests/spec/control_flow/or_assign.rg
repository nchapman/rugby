#@ run-pass
#@ check-output
#
# Test: Section 9.1 - ||= operator for optionals

# ||= assigns if absent
name : String? = nil
name ||= "Anonymous"
puts name

# ||= doesn't reassign if present
name2 : String? = "Alice"
name2 ||= "Anonymous"
puts name2

# ||= with empty string (present, not replaced)
name3 : String? = ""
name3 ||= "Anonymous"
puts "empty" if name3 == ""

#@ expect:
# Anonymous
# Alice
# empty
