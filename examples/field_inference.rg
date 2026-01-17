# Field Inference - How Rugby infers class fields
# Demonstrates: explicit declarations, parameter promotion, assignment inference

class User
  @role: String  # 1. Explicit declaration

  # 2. Parameter promotion (@name) - declares field and assigns argument
  # 3. Assignment inference (@email) - first assignment in initialize
  def initialize(@name: String, email: String, role: String)
    @email = email
    @role = role
  end

  def update_email(new_email: String)
    @email = new_email  # OK: @email already exists
  end

  def introduce
    puts "I'm #{@name}, email: #{@email}, role: #{@role}"
  end
end

u = User.new("Alice", "alice@example.com", "admin")
u.introduce
u.update_email("alice.new@example.com")
u.introduce
