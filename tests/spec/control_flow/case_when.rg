#@ run-pass
#@ check-output
#
# Test case/when statements with value matching
# TODO: case/when should implicitly return values (Ruby expression semantics)
#       Currently requires explicit `return` in each branch

def grade_to_letter(score : Int) -> String
  case score
  when 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100
    return "A"
  when 80, 81, 82, 83, 84, 85, 86, 87, 88, 89
    return "B"
  when 70, 71, 72, 73, 74, 75, 76, 77, 78, 79
    return "C"
  when 60, 61, 62, 63, 64, 65, 66, 67, 68, 69
    return "D"
  else
    return "F"
  end
end

def day_type(day : String) -> String
  case day
  when "Saturday", "Sunday"
    return "weekend"
  when "Monday", "Tuesday", "Wednesday", "Thursday", "Friday"
    return "weekday"
  else
    return "unknown"
  end
end

# Test numeric case/when
puts grade_to_letter(95)
puts grade_to_letter(85)
puts grade_to_letter(75)
puts grade_to_letter(65)
puts grade_to_letter(55)

# Test string case/when
puts day_type("Saturday")
puts day_type("Monday")
puts day_type("Holiday")

#@ expect:
# A
# B
# C
# D
# F
# weekend
# weekday
# unknown
