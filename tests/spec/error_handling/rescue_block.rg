#@ run-pass
#@ check-output
#
# Test: Section 16.5 - Block rescue with error binding

import "errors"

def fetch_data(url : String) -> (String, Error)
  if url == ""
    return "", errors.new("empty url")
  end
  "data from #{url}", nil
end

# Block form with error binding
data = fetch_data("") rescue => err do
  puts "Warning: #{err}"
  "default data"
end

puts data

#@ expect:
# Warning: empty url
# default data
