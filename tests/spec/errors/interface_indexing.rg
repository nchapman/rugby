#@ compile-fail
#
# BUG-048: Interface indexing
# Cannot index values of interface type 'any'

def get_post -> any
  { "title" => "Hello" }
end

post = get_post
title = post["title"]  #~ cannot index
