#@ run-pass
#@ check-output
#
# Test: Section 7.1 - Basic enums
# TODO: Implement enum syntax

enum Status
  Pending
  Active
  Completed
  Cancelled
end

status = Status::Active

case status
when Status::Pending
  puts "waiting"
when Status::Active
  puts "running"
else
  puts "done"
end

#@ expect:
# running
