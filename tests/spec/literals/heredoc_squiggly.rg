#@ run-pass
#@ check-output
#
# Test: Section 4.5 - Squiggly heredoc (<<~) strips leading whitespace

def get_query -> String
  <<~SQL
    SELECT *
    FROM users
    WHERE active = true
  SQL
end

query = get_query
puts query

#@ expect:
# SELECT *
# FROM users
# WHERE active = true
#
