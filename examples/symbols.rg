# Symbol literals demonstration
# Symbols are lightweight identifiers that compile to strings

def check_status(status)
  if status == :ok
    puts "Everything is working!"
  elsif status == :error
    puts "Something went wrong"
  elsif status == :pending
    puts "Still processing..."
  else
    puts "Unknown status"
  end
end

def main
  # Basic symbol usage
  current = :ok
  puts current

  # Symbols in arrays
  states = [:pending, :active, :completed]
  puts "Available states:"
  states.each do |s|
    puts s
  end

  # Symbols as function arguments
  check_status :ok
  check_status :error
  check_status :pending
  check_status :unknown

  # Symbol comparison
  if :success == :success
    puts "Symbols compare correctly"
  end
end
