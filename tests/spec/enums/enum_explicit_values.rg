#@ run-pass
#@ check-output
#
# Test: Section 7.2 - Enums with explicit values
# TODO: Implement enum syntax with values

enum HttpStatus
  Ok = 200
  Created = 201
  BadRequest = 400
  NotFound = 404
  ServerError = 500
end

status = HttpStatus::NotFound
puts status.value

#@ expect:
# 404
