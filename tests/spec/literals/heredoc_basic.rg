#@ run-pass
#@ check-output
#
# Test: Section 4.5 - Basic heredocs

# Basic heredoc
text = <<END
First line
Second line
END
puts text
puts "---"

# Heredoc with different delimiter
sql = <<SQL
SELECT *
FROM users
SQL
puts sql

#@ expect:
# First line
# Second line
#
# ---
# SELECT *
# FROM users
#
