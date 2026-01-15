#@ run-pass
#@ check-output
#
# Test: Section 16.3 - Custom error types

class ValidationError
  def initialize(@field : String, @message : String)
  end

  def error -> String
    "#{@field}: #{@message}"
  end
end

class NotFoundError
  def initialize(@resource : String, @id : Int)
  end

  def error -> String
    "#{@resource} not found: #{@id}"
  end
end

def validate_age(age : Int) -> Error
  if age < 0
    return ValidationError.new("age", "must be non-negative")
  end
  if age > 150
    return ValidationError.new("age", "must be realistic")
  end
  nil
end

err = validate_age(-5)
if err != nil
  puts err.error
end

err2 = validate_age(200)
if err2 != nil
  puts err2.error
end

err3 = validate_age(30)
if err3 == nil
  puts "valid"
end

#@ expect:
# age: must be non-negative
# age: must be realistic
# valid
