# Rugby Optionals
# Demonstrates: T?, nil, ??, &., if let, ok?, unwrap
#
# Prefer optional operators over manual nil checks for cleaner code.

def find_user(id: Int): String?
  return "Alice" if id == 1
  return "Bob" if id == 2
  nil
end

class Address
  getter city: String

  def initialize(@city: String, @zip: String)
  end
end

class User
  getter address: Address?

  def initialize(@name: String, @address: Address?)
  end
end

def main
  # Nil coalescing with ?? - the idiomatic way
  name = find_user(1) ?? "Anonymous"
  puts "User 1: #{name}"

  missing = find_user(99) ?? "Anonymous"
  puts "User 99: #{missing}"

  # Safe navigation with &.
  addr = Address.new("NYC", "10001")
  user_with_addr = User.new("Charlie", addr)
  user_no_addr = User.new("Dave", nil)

  city1 = user_with_addr.address&.city ?? "Unknown"
  city2 = user_no_addr.address&.city ?? "Unknown"
  puts "Charlie's city: #{city1}"
  puts "Dave's city: #{city2}"

  # if let - scoped binding (preferred for conditional use)
  if let user = find_user(2)
    puts "Found: #{user}"
  end

  if let missing = find_user(100)
    puts "This won't print"
  else
    puts "User 100 not found"
  end

  # Tuple unpacking - when you need the bool separately
  val, ok = find_user(1)
  if ok
    puts "Tuple: found #{val}"
  end

  # Optional methods (present?/absent? are synonyms for ok?/nil?)
  result = find_user(1)
  puts "ok? #{result.ok?}, nil? #{result.nil?}"
  puts "present? #{result.present?}, absent? #{result.absent?}"

  # map on optional - transform if present
  upper = find_user(1).map { |s| s.upcase }
  puts "Mapped: #{upper ?? "none"}"

  # unwrap - only when absence is a bug (panics if nil)
  known = find_user(1).unwrap
  puts "Unwrapped: #{known}"
end
