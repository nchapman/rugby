# Case statements in Rugby
# Demonstrates case/when/else syntax that compiles to Go switch

def describe_status(status : Int)
  case status
  when 200
    puts "OK"
  when 201
    puts "Created"
  when 400
    puts "Bad Request"
  when 404
    puts "Not Found"
  when 500, 502, 503
    puts "Server Error"
  else
    puts "Unknown"
  end
end

def grade(score : Int)
  # Case without subject (like switch true)
  case
  when score >= 90
    puts "A"
  when score >= 80
    puts "B"
  when score >= 70
    puts "C"
  when score >= 60
    puts "D"
  else
    puts "F"
  end
end

def main
  # Test status codes
  puts "Status descriptions:"
  describe_status(200)
  describe_status(404)
  describe_status(502)
  describe_status(999)

  puts ""

  # Test grades
  puts "Grades:"
  grade(95)
  grade(85)
  grade(75)
  grade(55)
end
