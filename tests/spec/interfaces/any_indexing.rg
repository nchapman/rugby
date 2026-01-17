#@ run-pass
#@ check-output
#
# Test that values of type 'Any' can be indexed

def get_post: Any
  { "title" => "Hello", "body" => "World" }
end

post = get_post
title = post["title"]
body = post["body"]

puts title
puts body

#@ expect:
# Hello
# World
