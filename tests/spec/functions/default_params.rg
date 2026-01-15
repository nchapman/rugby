#@ compile-fail
#@ skip: Default parameters not yet implemented (Section 10.3)
#
# Test: Section 10.3 - Default parameters
# TODO: Implement default parameter syntax

def connect(host : String, port : Int = 8080, timeout : Int = 30)
  puts "#{host}:#{port} timeout=#{timeout}"
end

connect("localhost")
connect("localhost", 3000)
connect("localhost", 3000, 60)
