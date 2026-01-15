#@ run-pass
#@ check-output
#
# Test: Section 15.2 - Class visibility levels

class User
  pub getter name : String

  def initialize(@name : String, @password : String)
  end

  # Public - accessible from anywhere (with pub)
  pub def display -> String
    "User: #{@name}"
  end

  # Package-private (default) - accessible within same package
  def validate -> Bool
    @password.length >= 8
  end

  # Private - accessible only within this class
  private def encrypt_password -> String
    "***"
  end

  pub def save
    if validate
      encrypted = encrypt_password
      puts "Saved #{@name}"
    end
  end
end

user = User.new("Alice", "password123")
puts user.display
user.save

#@ expect:
# User: Alice
# Saved Alice
