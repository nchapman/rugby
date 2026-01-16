#@ run-pass
#@ check-output
#
# Integration Test: Data Processing Pipeline
# Features: Classes + Arrays + Lambdas + Iteration methods + Symbol-to-proc

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
engineers = people.select -> (p) { p.department == "Engineering" && p.age > 27 }
engineer_names = engineers.map(&:name)
engineer_names.each -> (name) { puts "  #{name}" }

# Pipeline 2: Calculate total age of marketing team
puts "Marketing team:"
marketing = people.select -> (p) { p.department == "Marketing" }
puts "  Count: #{marketing.length}"
sum = marketing.reduce(0) -> (acc, p) { acc + p.age }
puts "  Total age: #{sum}"
puts "  Average age: #{sum / marketing.length}"

# Pipeline 3: Count by department
puts "Department counts:"
departments = people.map(&:department)
eng_count = departments.select -> (d) { d == "Engineering" }.length
mkt_count = departments.select -> (d) { d == "Marketing" }.length
puts "  Engineering: #{eng_count}"
puts "  Marketing: #{mkt_count}"

# Pipeline 4: Check conditions with any?/all?
puts "Team checks:"
has_senior = people.any? -> (p) { p.age >= 35 }
all_adults = people.all? -> (p) { p.age >= 18 }
none_retired = people.none? -> (p) { p.age >= 65 }
puts "  Has senior (35+): #{has_senior}"
puts "  All adults (18+): #{all_adults}"
puts "  None retired (65+): #{none_retired}"

# Pipeline 5: Find specific person
puts "Find operations:"
carol = people.find -> (p) { p.name == "Carol" }
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
