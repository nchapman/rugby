#@ run-pass
#@ check-output
#
# Test: Section 11.2 - Instance variables

class User
  @role: String                # explicit declaration

  def initialize(@name: String) # parameter promotion
    @age = 0                     # inferred from initialize
    @role = "user"
  end

  def info: String
    "#{@name}, #{@age}, #{@role}"
  end

  def set_age(age: Int)
    @age = age
  end
end

user = User.new("Alice")
puts user.info

user.set_age(30)
puts user.info

#@ expect:
# Alice, 0, user
# Alice, 30, user
