#@ run-pass
#@ check-output
#
# Test symbol literals (compile to strings in Go)

# Basic symbols
status = :ok
action = :create

puts status
puts action

# Symbols in conditionals
if status == :ok
  puts "success"
end

# Symbols as map keys
config = {
  mode: :production,
  debug: false
}
puts config["mode"]

#@ expect:
# ok
# create
# success
# production
