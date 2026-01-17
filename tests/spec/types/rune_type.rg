#@ run-pass
#@ check-output

# Test Rune type - represents a Unicode code point

# Function returning Rune
def get_char_code: Rune
  65  # ASCII 'A'
end

# Variable with Rune type annotation
r: Rune = get_char_code()
puts r

# Rune as function parameter
def print_code(c: Rune)
  puts c
end

print_code(97)  # ASCII 'a'

# Rune in function returning Rune
def echo_rune(c: Rune): Rune
  c
end

r2 = echo_rune(72)  # 'H'
puts r2

#@ expect:
# 65
# 97
# 72
