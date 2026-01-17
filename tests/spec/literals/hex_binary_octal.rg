#@ run-pass
#@ check-output

# Test hex, binary, and octal integer literals

def main
  # Hex literals
  puts 0x80      # 128
  puts 0xFF      # 255
  puts 0xABCD    # 43981

  # Binary literals
  puts 0b1010    # 10
  puts 0b11111111 # 255

  # Octal literals
  puts 0o755     # 493
  puts 0o17      # 15

  # Use in expressions with shift (already working)
  puts 0x80 >> 4  # 8
  puts 0b1000 << 1  # 16
end

#@ expect:
# 128
# 255
# 43981
# 10
# 255
# 493
# 15
# 8
# 16
