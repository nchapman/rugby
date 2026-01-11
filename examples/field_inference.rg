class User
  @role : String  # 1. explicit declaration

  def initialize(@name : String, email : String)  # 2. parameter promotion (@name)
    @email = email  # 3. first assignment inference
  end

  def update_email(new_email : String)
    @email = new_email  # OK: @email already exists
  end

  def introduce
    puts("I'm #{@name}, email: #{@email}")
  end
end

u = User.new("Alice", "alice@example.com")
u.introduce()
