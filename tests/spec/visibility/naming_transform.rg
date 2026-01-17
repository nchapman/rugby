#@ run-pass
#@ check-output
#
# Test: Section 15.4 - Naming transformations (snake_case to camelCase)

# These names will be transformed in generated Go:
# find_user -> findUser (internal)
# pub find_user -> FindUser (exported)
# user_id -> userID (acronym handling)

def find_user(user_id: Int): String
  "User #{user_id}"
end

pub def get_user_by_id(user_id: Int): String
  find_user(user_id)
end

# Test that it compiles and runs correctly
puts get_user_by_id(123)
puts find_user(456)

#@ expect:
# User 123
# User 456
