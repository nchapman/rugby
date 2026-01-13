# Rugby Control Flow
# Demonstrates: if/elsif/else, unless, case/when, case_type, ternary (spec 5.2)

def main
  # Basic if/elsif/else
  score = 85
  if score >= 90
    puts "Grade: A"
  elsif score >= 80
    puts "Grade: B"
  elsif score >= 70
    puts "Grade: C"
  else
    puts "Grade: F"
  end

  # unless (inverse of if)
  empty = false
  unless empty
    puts "Not empty!"
  end

  # Statement modifier form (spec 5.5)
  value = 10
  puts "Positive" if value > 0
  puts "Not zero" unless value == 0

  # Ternary operator (spec 5.2)
  max = 5 > 3 ? 5 : 3
  puts "Max: #{max}"

  # case/when with single values
  status = 200
  case status
  when 200
    puts "OK"
  when 404
    puts "Not Found"
  when 500
    puts "Server Error"
  else
    puts "Unknown"
  end

  # case/when with multiple values
  day = 6
  case day
  when 0, 6
    puts "Weekend"
  when 1, 2, 3, 4, 5
    puts "Weekday"
  end

  # case without subject (like switch true)
  temp = 25
  case
  when temp < 0
    puts "Freezing"
  when temp < 20
    puts "Cold"
  when temp < 30
    puts "Nice"
  else
    puts "Hot"
  end

  # case_type for type matching (spec 5.2)
  describe_type("hello")
  describe_type(42)
  describe_type(3.14)
end

# case_type matches on type (spec 5.2)
def describe_type(x : any)
  case_type x
  when String
    puts "String: #{x}"
  when Int
    puts "Int: #{x}"
  when Float
    puts "Float: #{x}"
  else
    puts "Unknown type"
  end
end
