# HTTP and JSON - Fetching data from an API
# Demonstrates: rugby/http stdlib, error handling with ! and rescue

import rugby/http

# Fetch a user from JSONPlaceholder API
def fetch_user(id : Int) -> (Map[String, any], error)
  url = "https://jsonplaceholder.typicode.com/users/#{id}"
  resp = http.get(url)!
  resp.json
end

# Fetch posts for a user
def fetch_posts(user_id : Int) -> (Array[any], error)
  url = "https://jsonplaceholder.typicode.com/users/#{user_id}/posts"
  resp = http.get(url)!
  resp.json_array
end

def main
  puts "Fetching user 1..."

  user = fetch_user(1) rescue do
    puts "Failed to fetch user"
    return
  end

  puts "Name: #{user["name"]}"
  puts "Email: #{user["email"]}"
  puts "Company: #{user["company"]}"

  puts "\nFetching posts..."

  posts = fetch_posts(1) rescue do
    puts "Failed to fetch posts"
    return
  end

  puts "Found #{posts.length} posts"

  # Show first 3 posts (use slice to avoid out-of-bounds)
  posts[0..2].each do |post|
    puts "- #{post["title"]}"
  end
end
