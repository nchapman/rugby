class User
  @role : String  # 1. explicit declaration

  def initialize(@name : String, email : String, role : String)  # 2. parameter promotion (@name)
    @email = email  # 3. first assignment inference
    @role = role
  end

  def update_email(new_email : String)
    @email = new_email  # OK: @email already exists
  end

  def introduce
    puts("I'm #{@name}, email: #{@email}, role: #{@role}")
  end
end

u = User.new("Alice", "alice@example.com", "admin")
u.introduce()
u.update_email("alice.new@example.com")
u.introduce()
