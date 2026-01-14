#@ run-pass
#@ check-output
#
# Test heredoc string literals

# Basic heredoc
text = <<END
Hello
World
END
puts text

# Heredoc with different delimiter
message = <<MSG
Line one
Line two
MSG
puts message

# Empty heredoc
empty = <<DONE
DONE
puts "empty length:"
puts empty.length

#@ expect:
# Hello
# World
# Line one
# Line two
# empty length:
# 0
