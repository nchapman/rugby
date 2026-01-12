# Example: HTTP and JSON stdlib usage
# Demonstrates fetching data from an API and parsing JSON

import rugby/http

# Fetch a user from JSONPlaceholder API
def fetch_user(id : Int) -> (Map[String, any], error)
  url = "https://jsonplaceholder.typicode.com/users/#{id}"
  resp = http.Get(url)!

  if !resp.Ok()
    return nil, fmt.Errorf("HTTP error: %d", resp.Status)
  end

  resp.JSON()
end

# Fetch posts for a user
def fetch_posts(user_id : Int) -> (Array[any], error)
  url = "https://jsonplaceholder.typicode.com/users/#{user_id}/posts"
  resp = http.Get(url)!

  if !resp.Ok()
    return nil, fmt.Errorf("HTTP error: %d", resp.Status)
  end

  resp.JSONArray()
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

  # Show first 3 posts
  for i in 0..2
    puts "- #{posts[i]["title"]}"
  end
end
