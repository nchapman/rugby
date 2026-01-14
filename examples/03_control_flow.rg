# Rugby Control Flow
# Demonstrates: if/elsif/else, unless, case/when, case_type, ternary, modifiers

def main
  # if/elsif/else
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

  # Statement modifiers - put simple conditions at the end
  value = 10
  puts "Positive" if value > 0
  puts "Not zero" unless value == 0

  # Ternary operator
  max = 5 > 3 ? 5 : 3
  puts "Max: #{max}"

  # case/when - value matching
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

  # case_type - type matching (Go type switch)
  describe_type("hello")
  describe_type(42)
  describe_type(3.14)
end

def describe_type(x : any)
  case_type x
  when String
    puts "It's a string: #{x}"
  when Int
    puts "It's an int: #{x}"
  when Float
    puts "It's a float: #{x}"
  else
    puts "Unknown type"
  end
end
