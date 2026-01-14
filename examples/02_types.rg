# Rugby Type System
# Demonstrates: primitives, type inference, arrays, maps, symbols

def main
  # Type inference - Rugby infers types from values
  n = 42              # Int
  pi = 3.14159        # Float
  flag = true         # Bool
  greeting = "hello"  # String

  puts "Int: #{n}, Float: #{pi}, Bool: #{flag}"

  # Explicit type annotations when needed
  count : Int = 100
  ratio : Float = 2.5

  # Empty typed arrays
  empty_nums : Array[Int] = []
  empty_strs = [] : Array[String]  # inline annotation syntax

  # Symbols - compile to strings, great for status values
  status = :active
  puts "Status: #{status}"

  # Arrays
  nums = [1, 2, 3, 4, 5]
  puts "Array: #{nums}"
  puts "First: #{nums[0]}, Last: #{nums[-1]}, Length: #{nums.length}"

  # Word arrays with %w{}
  fruits = %w{apple banana cherry}
  puts "Fruits: #{fruits}"

  # Interpolated word arrays with %W{}
  lang = "Rugby"
  phrases = %W{hello #{lang} world}
  puts "Phrases: #{phrases}"

  # Array splat
  rest = [2, 3]
  all = [1, *rest, 4]
  puts "Splat: #{all}"

  # Maps - hash rocket or symbol shorthand
  ages = {"alice" => 30, "bob" => 25}
  puts "Alice's age: #{ages["alice"]}"

  # Symbol shorthand and implicit values
  host = "localhost"
  port = 8080
  config = {host:, port:}  # same as {host: host, port: port}
  puts "Config: #{config}"

  # Double-splat merges maps
  defaults = {timeout: 30, retries: 3}
  overrides = {timeout: 60}
  merged = {**defaults, **overrides}
  puts "Merged: #{merged}"

  # Integer methods
  x = -7
  puts "#{x}.abs = #{x.abs}, even? = #{x.even?}, odd? = #{x.odd?}"

  # Float methods
  f = 3.7
  puts "#{f}.floor = #{f.floor}, ceil = #{f.ceil}, round = #{f.round}"

  # Type conversions
  puts "42.to_s = #{42.to_s}, 3.to_f = #{3.to_f}"
end
