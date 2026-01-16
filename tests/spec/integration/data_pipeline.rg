#@ run-pass
#@ check-output
#
# Integration Test: Data Processing Pipeline
# Features: Classes + Arrays + Blocks + Lambdas + Iteration methods
# Note: Uses explicit type annotations in lambdas (required for getter inference)

# Data record class
class Person
  def initialize(@name : String, @age : Int, @department : String)
  end

  getter name : String
  getter age : Int
  getter department : String
end

# Create sample data
people = [
  Person.new("Alice", 30, "Engineering"),
  Person.new("Bob", 25, "Marketing"),
  Person.new("Carol", 35, "Engineering"),
  Person.new("Dave", 28, "Engineering"),
  Person.new("Eve", 32, "Marketing")
]

# Pipeline 1: Filter engineers over 27
puts "Engineers over 27:"
engineers = people.select -> (p : Person) { p.department == "Engineering" && p.age > 27 }
engineer_names = engineers.map -> (p : Person) { p.name }
engineer_names.each do |name|
  puts "  #{name}"
end

# Pipeline 2: Calculate total age of marketing team
puts "Marketing team:"
marketing = people.select -> (p : Person) { p.department == "Marketing" }
puts "  Count: #{marketing.length}"
# Calculate sum using each
sum = 0
marketing.each do |p|
  sum = sum + p.age
end
puts "  Total age: #{sum}"
puts "  Average age: #{sum / marketing.length}"

# Pipeline 3: Count by department
puts "Department counts:"
departments = people.map -> (p : Person) { p.department }
eng_depts = departments.select -> (d : String) { d == "Engineering" }
mkt_depts = departments.select -> (d : String) { d == "Marketing" }
puts "  Engineering: #{eng_depts.length}"
puts "  Marketing: #{mkt_depts.length}"

# Pipeline 4: Check conditions with any?/all?
puts "Team checks:"
has_senior = people.any? -> (p : Person) { p.age >= 35 }
all_adults = people.all? -> (p : Person) { p.age >= 18 }
none_retired = people.none? -> (p : Person) { p.age >= 65 }
puts "  Has senior (35+): #{has_senior}"
puts "  All adults (18+): #{all_adults}"
puts "  None retired (65+): #{none_retired}"

# Pipeline 5: Find specific person
puts "Find operations:"
carol = people.find -> (p : Person) { p.name == "Carol" }
if let found = carol
  puts "  Found Carol, age #{found.age}"
end

#@ expect:
# Engineers over 27:
#   Alice
#   Carol
#   Dave
# Marketing team:
#   Count: 2
#   Total age: 57
#   Average age: 28
# Department counts:
#   Engineering: 3
#   Marketing: 2
# Team checks:
#   Has senior (35+): true
#   All adults (18+): true
#   None retired (65+): true
# Find operations:
#   Found Carol, age 35
