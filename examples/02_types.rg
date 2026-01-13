# Rugby Type System
# Demonstrates: primitives, type inference, annotations, composites (spec 4)

def main
  # Primitive types with inference (spec 4.1)
  n = 42                    # Int
  pi = 3.14159              # Float
  flag = true               # Bool
  greeting = "hello"        # String

  puts "Int: #{n}"
  puts "Float: #{pi}"
  puts "Bool: #{flag}"
  puts "String: #{greeting}"

  # Explicit type annotations (spec 4.3)
  count : Int = 100
  ratio : Float = 2.5
  puts "Annotated: #{count}, #{ratio}"

  # Symbols compile to strings (spec 4.1.1)
  status = :active
  puts "Symbol: #{status}"

  # Arrays (spec 4.2)
  nums = [1, 2, 3, 4, 5]
  puts "Array: #{nums}"
  puts "First: #{nums[0]}, Last: #{nums[-1]}"
  puts "Length: #{nums.length}"

  # Word arrays with %w (spec 4.2)
  words = %w{apple banana cherry}
  puts "Words: #{words}"

  # Word arrays with interpolation %W (spec 4.2)
  lang = "Rugby"
  phrases = %W{hello #{lang} world}
  puts "Interpolated words: #{phrases}"

  # Array splat (spec 4.2)
  rest = [2, 3]
  all = [1, *rest, 4]
  puts "Splat array: #{all}"

  # Maps with hash rocket (spec 4.2)
  ages = {"alice" => 30, "bob" => 25}
  puts "Alice's age: #{ages["alice"]}"

  # Maps with symbol shorthand (spec 4.2)
  host = "localhost"
  port = "8080"
  config = {host: host, port: port}
  puts "Config: #{config}"

  # Map double-splat merge (spec 4.2)
  defaults = {timeout: "30", retries: "3"}
  overrides = {timeout: "60"}
  merged = {**defaults, **overrides}
  puts "Merged config: #{merged}"

  # Integer methods (spec 12.6)
  x = -7
  puts "#{x}.even? = #{x.even?}"
  puts "#{x}.odd? = #{x.odd?}"
  puts "#{x}.abs = #{x.abs}"
  puts "#{x}.positive? = #{x.positive?}"

  # Float methods (spec 12.7)
  f = 3.7
  puts "#{f}.floor = #{f.floor}"
  puts "#{f}.ceil = #{f.ceil}"
  puts "#{f}.round = #{f.round}"

  # Type conversions
  num = 42
  puts "to_s: #{num.to_s}"
  puts "to_f: #{num.to_f}"
end
