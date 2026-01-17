#@ run-pass
#@ check-output
#
# Test: Section 14.3 - Module namespacing

module Http
  class Response
    getter status: Int
    getter body: String

    def initialize(@status: Int, @body: String)
    end
  end

  def self.get(url: String): Response
    Response.new(200, "Hello from #{url}")
  end
end

response = Http.get("https://example.com")
puts response.status
puts response.body

# Using fully qualified name
response2 = Http::Response.new(404, "Not Found")
puts response2.status

#@ expect:
# 200
# Hello from https://example.com
# 404
