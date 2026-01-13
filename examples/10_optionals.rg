# Rugby Optionals
# Demonstrates: T?, nil, ??, &., if let, ok?, nil?, unwrap (spec 4.4)

# Function returning optional
def find_user(id : Int) -> String?
  return "Alice" if id == 1
  return "Bob" if id == 2
  nil
end

class Address
  @city : String
  @zip : String

  def initialize(@city : String, @zip : String)
  end

  def city -> String
    @city
  end
end

class User
  @name : String
  @address : Address?

  def initialize(@name : String, @address : Address?)
  end

  def address -> Address?
    @address
  end
end

def main
  # Basic optional usage (spec 4.4.2)
  result = find_user(1)
  if result.ok?
    puts "Found: #{result.unwrap}"
  end

  result2 = find_user(99)
  if result2.nil?
    puts "User 99 not found"
  end

  # Nil coalescing with ?? (spec 4.4.1)
  name1 = find_user(1) ?? "Anonymous"
  name2 = find_user(99) ?? "Anonymous"
  puts "User 1: #{name1}"
  puts "User 99: #{name2}"

  # if let - scoped binding (spec 4.4.3)
  if let user = find_user(2)
    puts "if let found: #{user}"
  end

  if let missing = find_user(100)
    puts "This won't print"
  else
    puts "if let else: not found"
  end

  # Tuple unpacking - Go's comma-ok idiom (spec 4.4.3)
  val, ok = find_user(1)
  if ok
    puts "Tuple ok: #{val}"
  end

  # Safe navigation with &. (spec 4.4.1)
  addr = Address.new("NYC", "10001")
  user_with_addr = User.new("Charlie", addr)
  user_no_addr = User.new("Dave", nil)

  city1 = user_with_addr.address&.city ?? "Unknown"
  city2 = user_no_addr.address&.city ?? "Unknown"
  puts "Charlie's city: #{city1}"
  puts "Dave's city: #{city2}"

  # Optional methods (spec 4.4.2)
  opt = find_user(1)
  puts "present? #{opt.present?}"
  puts "absent? #{opt.absent?}"

  # map on optional
  upper = find_user(1).map { |s| s.upcase }
  if upper.ok?
    puts "Mapped: #{upper.unwrap}"
  end

  # each on optional (execute if present)
  find_user(2).each do |name|
    puts "each: found #{name}"
  end
end
