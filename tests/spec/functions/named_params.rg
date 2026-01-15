#@ compile-fail
#@ skip: Named parameters not yet implemented (Section 10.4)
#
# Test: Section 10.4 - Named parameters
# TODO: Implement named parameter syntax

def fetch(url : String, headers: Map<String, String> = {},
          timeout: Int = 30, retry: Bool = true)
  puts url
end

fetch("https://example.com")
fetch("https://example.com", timeout: 60)
fetch("https://example.com", retry: false, timeout: 120)
