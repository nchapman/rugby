# HTTP and JSON - Fetching data from an API
# Demonstrates: rugby/http stdlib, error handling with ! and rescue

import rugby/http

# Fetch a user from JSONPlaceholder API
def fetch_user(id: Int): (Map<String, any>, Error)
  url = "https://jsonplaceholder.typicode.com/users/#{id}"
  resp = http.get(url)!
  resp.json
end

# Fetch posts for a user
def fetch_posts(user_id: Int): (Array<any>, Error)
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

  # Show first 3 post titles (using times loop to avoid range slice limitation)
  3.times -> do |i|
    if i < posts.length
      puts "- #{posts[i]["title"]}"
    end
  end
end
