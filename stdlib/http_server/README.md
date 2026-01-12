# rugby/http_server

A fast, ergonomic HTTP server for building web applications and APIs in Rugby.

## Import

```ruby
import rugby/http_server
```

## Quick Start

```ruby
import rugby/http_server

app = http_server.new()

app.get "/" do |ctx|
  ctx.text("Hello, World!")
end

app.get "/users/:id" do |ctx|
  id = ctx.param("id")
  ctx.json({ id: id, name: "Alice" })
end

app.post "/users" do |ctx|
  user = ctx.json()!
  ctx.status(201)
  ctx.json(user)
end

app.listen(":8080")!
```

## Routing

### HTTP Methods

```ruby
app.get    "/path" do |ctx| ... end
app.post   "/path" do |ctx| ... end
app.put    "/path" do |ctx| ... end
app.patch  "/path" do |ctx| ... end
app.delete "/path" do |ctx| ... end
app.head   "/path" do |ctx| ... end
app.options "/path" do |ctx| ... end
```

### Path Parameters

Use `:name` for named parameters:

```ruby
app.get "/users/:id" do |ctx|
  id = ctx.param("id")
  ctx.json({ id: id })
end

app.get "/posts/:post_id/comments/:id" do |ctx|
  post_id = ctx.param("post_id")
  comment_id = ctx.param("id")
  # ...
end
```

### Wildcards

Use `*name` to capture the rest of the path:

```ruby
app.get "/files/*path" do |ctx|
  path = ctx.param("path")  # "images/logo.png" for /files/images/logo.png
  ctx.file("./static/#{path}")!
end
```

## Routes

Routes allow you to organize handlers across multiple files and mount them with URL prefixes.

### Defining Routes

```ruby
# routes/users.rg
import rugby/http_server

pub def users_routes() -> http_server.Routes
  r = http_server.routes()

  r.get "/" do |ctx|
    users = all_users()!
    ctx.json(users)
  end

  r.get "/:id" do |ctx|
    id = ctx.param("id").to_i!
    user = find_user(id)
    if user.nil?
      ctx.status(404)
      ctx.json({ error: "user not found" })
      return
    end
    ctx.json(user)
  end

  r.post "/" do |ctx|
    data = ctx.json(CreateUserRequest)!
    user = create_user(data)!
    ctx.status(201)
    ctx.json(user)
  end

  r.put "/:id" do |ctx|
    id = ctx.param("id").to_i!
    data = ctx.json(UpdateUserRequest)!
    user = update_user(id, data)!
    ctx.json(user)
  end

  r.delete "/:id" do |ctx|
    id = ctx.param("id").to_i!
    delete_user(id)!
    ctx.status(204)
  end

  r
end
```

### Mounting Routes

```ruby
# main.rg
import rugby/http_server
import "./routes/users"
import "./routes/posts"

app = http_server.new()

app.mount "/users", users_routes()
app.mount "/posts", posts_routes()

app.listen(":8080")!
```

### Nested Routes

Use block syntax for nested mounting with shared middleware:

```ruby
app.mount "/api/v1" do |api|
  api.use rate_limit(100)

  api.mount "/users", users_routes()
  api.mount "/posts", posts_routes()

  api.mount "/admin" do |admin|
    admin.use require_role(:admin)
    admin.mount "/stats", admin_stats_routes()
    admin.mount "/users", admin_users_routes()
  end
end
```

## Context

The `ctx` parameter provides access to request data and response methods.

### Request Data

```ruby
# Path and method
ctx.method                     # "GET", "POST", etc.
ctx.path                       # "/users/123"

# Path parameters
ctx.param("id")                # path parameter -> String

# Query string
ctx.query("page")              # single param -> String?
ctx.query("page") ?? "1"       # with default
ctx.queries("tags")            # multi-value -> Array[String]
                               # e.g., ?tags=a&tags=b -> ["a", "b"]

# Headers
ctx.header("Authorization")    # single header -> String?
ctx.header("Content-Type")     # -> String?
ctx.headers                    # all headers -> Map[String, String]

# Client info
ctx.ip                         # client IP address -> String
```

### Request Body

All body parsing methods return errors - use `!` or `rescue`:

```ruby
# JSON body -> Map[String, any]
data = ctx.json()!
name = data["name"]

# JSON body -> typed struct
user = ctx.json(User)!

# Raw text body
text = ctx.text()!

# Form data (application/x-www-form-urlencoded)
form = ctx.form()!
username = form["username"]

# File upload (multipart/form-data)
file = ctx.form_file("avatar")!
file.filename    # original filename
file.size        # file size in bytes
file.content     # file contents as Bytes
file.save("./uploads/#{file.filename}")!
```

### Response Methods

```ruby
# Set status code (call before writing body)
ctx.status(201)
ctx.status(404)

# Set response header
ctx.set_header("X-Custom", "value")
ctx.set_header("Cache-Control", "max-age=3600")

# Send JSON (sets Content-Type: application/json)
ctx.json({ name: "Alice", age: 30 })
ctx.json(user)                 # struct with json tags

# Send plain text (sets Content-Type: text/plain)
ctx.text("Hello, World!")

# Send HTML (sets Content-Type: text/html)
ctx.html("<h1>Hello</h1>")

# Send file
ctx.file("./report.pdf")!
ctx.file("./image.png")!       # auto-detects content type

# Redirect
ctx.redirect("/login")         # 302 Found
ctx.redirect("/new-url", 301)  # 301 Moved Permanently

# Streaming response
ctx.stream do |w|
  w.write("chunk 1")!
  w.write("chunk 2")!
  w.flush()!
end
```

### Cookies

```ruby
# Read cookie
session = ctx.cookie("session")  # -> String?

if let token = ctx.cookie("auth_token")
  # use token
end

# Set cookie
ctx.set_cookie("session", token)

# Set cookie with options
ctx.set_cookie("session", token, {
  http_only: true,
  secure: true,
  same_site: :strict,          # :strict, :lax, or :none
  max_age: 86400,              # seconds
  path: "/",
  domain: "example.com"
})

# Delete cookie
ctx.delete_cookie("session")
```

### Middleware Data

Pass data between middleware and handlers:

```ruby
# In middleware - set value (use symbols as keys)
ctx.set(:user, current_user)
ctx.set(:request_id, uuid.v4())

# In handler - get value
user = ctx.get(:user)          # -> any?

if let user = ctx.get(:user)
  # user is available
end

# Type assertion
user = ctx.get(:user).as(User) ?? guest_user
```

## Middleware

Middleware wraps handlers to add cross-cutting functionality.

### Creating Middleware

```ruby
pub def logger() -> http_server.Middleware
  http_server.middleware do |ctx, next_fn|
    start = time.now()
    next_fn.call()
    elapsed = time.since(start)
    puts "#{ctx.method} #{ctx.path} #{ctx.status_code} #{elapsed}"
  end
end

pub def require_auth() -> http_server.Middleware
  http_server.middleware do |ctx, next_fn|
    token = ctx.header("Authorization")
    if token.nil? || token.empty?
      ctx.status(401)
      ctx.json({ error: "unauthorized" })
      return
    end

    user = validate_token(token) rescue do
      ctx.status(401)
      ctx.json({ error: "invalid token" })
      return
    end

    ctx.set(:user, user)
    next_fn.call()
  end
end

pub def require_role(role : Symbol) -> http_server.Middleware
  http_server.middleware do |ctx, next_fn|
    user = ctx.get(:user).as(User)
    if user.nil? || user.role != role
      ctx.status(403)
      ctx.json({ error: "forbidden" })
      return
    end
    next_fn.call()
  end
end
```

### Using Middleware

```ruby
# Global middleware (applies to all routes)
app.use logger()
app.use http_server.recover()

# Route-level middleware
app.get "/admin", require_auth(), require_role(:admin) do |ctx|
  ctx.json({ message: "welcome admin" })
end

# Routes-level middleware
app.mount "/api" do |api|
  api.use require_auth()
  api.mount "/users", users_routes()
end
```

### Middleware Execution Order

Middleware executes in the order it's added:

```ruby
app.use first()    # runs first
app.use second()   # runs second
app.use third()    # runs third

# Request flow:  first -> second -> third -> handler
# Response flow: third <- second <- first <- handler
```

## Built-in Middleware

### http_server.logger(options?)

Logs requests with timing information.

```ruby
app.use http_server.logger()

# With options
app.use http_server.logger({
  format: :combined,           # :combined, :common, or :short
  skip: do |ctx|               # skip logging for certain requests
    ctx.path == "/health"
  end
})
```

Output: `GET /users/123 200 12.34ms`

### http_server.recover()

Recovers from panics and returns 500 error.

```ruby
app.use http_server.recover()

# With custom handler
app.use http_server.recover do |ctx, err|
  log.error("panic recovered", { error: err, path: ctx.path })
  ctx.status(500)
  ctx.json({ error: "internal server error" })
end
```

### http_server.request_id()

Adds a unique request ID to each request.

```ruby
app.use http_server.request_id()

# Access in handlers
app.get "/" do |ctx|
  id = ctx.get(:request_id)
  # Also sets X-Request-ID response header
end
```

### http_server.cors(options)

Handles Cross-Origin Resource Sharing.

```ruby
# Allow all origins
app.use http_server.cors({ origins: ["*"] })

# Specific configuration
app.use http_server.cors({
  origins: ["https://example.com", "https://app.example.com"],
  methods: ["GET", "POST", "PUT", "DELETE"],
  headers: ["Authorization", "Content-Type"],
  expose_headers: ["X-Total-Count"],
  credentials: true,
  max_age: 86400                # preflight cache in seconds
})
```

### http_server.compress()

Compresses responses with gzip/deflate.

```ruby
app.use http_server.compress()

# With options
app.use http_server.compress({
  level: 6,                    # compression level 1-9
  min_size: 1024               # minimum size to compress
})
```

### http_server.static(url_path, fs_path)

Serves static files from a directory.

```ruby
app.use http_server.static("/public", "./static")
app.use http_server.static("/assets", "./dist/assets")

# With options
app.use http_server.static("/public", "./static", {
  index: "index.html",         # serve index for directories
  max_age: 86400               # cache control max-age
})
```

### http_server.timeout(seconds)

Sets request timeout.

```ruby
app.use http_server.timeout(30)

# With custom handler
app.use http_server.timeout(30) do |ctx|
  ctx.status(504)
  ctx.json({ error: "request timeout" })
end
```

### http_server.rate_limit(options)

Rate limits requests by IP.

```ruby
app.use http_server.rate_limit({
  requests: 100,               # max requests
  window: 60                   # per window in seconds
})

# With custom key function
app.use http_server.rate_limit({
  requests: 100,
  window: 60,
  key: do |ctx|
    ctx.header("X-API-Key") ?? ctx.ip
  end
})
```

### http_server.basic_auth(options)

HTTP Basic Authentication.

```ruby
app.use http_server.basic_auth({
  user: "admin",
  pass: env.get("ADMIN_PASS")!
})

# With custom validator
app.use http_server.basic_auth do |user, pass|
  validate_credentials(user, pass)  # return Bool
end
```

### http_server.bearer_auth(validator)

Bearer token authentication.

```ruby
app.use http_server.bearer_auth do |token|
  user = validate_jwt(token)   # return User? or nil
  user                         # sets ctx[:user] if not nil
end
```

## Error Handling

### Custom Error Handlers

```ruby
# Handle specific status codes
app.on_error 404 do |ctx|
  ctx.json({ error: "not found", path: ctx.path })
end

app.on_error 500 do |ctx, err|
  log.error("internal error", { error: err, path: ctx.path })
  ctx.json({ error: "internal server error" })
end

# Catch-all error handler
app.on_error do |ctx, err|
  case ctx.status_code
  when 404
    ctx.json({ error: "not found" })
  when 401
    ctx.json({ error: "unauthorized" })
  when 403
    ctx.json({ error: "forbidden" })
  else
    ctx.json({ error: "something went wrong" })
  end
end
```

### Error Handling in Handlers

Use Rugby's standard error handling:

```ruby
app.get "/users/:id" do |ctx|
  # Parse with error propagation
  id = ctx.param("id").to_i!

  # Parse with fallback
  id = ctx.param("id").to_i rescue do
    ctx.status(400)
    ctx.json({ error: "invalid id" })
    return
  end

  # Handle not found
  user = find_user(id)
  if user.nil?
    ctx.status(404)
    ctx.json({ error: "user not found" })
    return
  end

  ctx.json(user)
end

app.post "/users" do |ctx|
  # Validate request body
  data = ctx.json(CreateUserRequest) rescue => err do
    ctx.status(400)
    ctx.json({ error: "invalid json", details: err.to_s })
    return
  end

  # Handle creation errors
  user = create_user(data) rescue => err do
    ctx.status(422)
    ctx.json({ error: "could not create user", details: err.to_s })
    return
  end

  ctx.status(201)
  ctx.json(user)
end
```

## Server Configuration

### Options

```ruby
app = http_server.new({
  read_timeout: 5,             # seconds
  write_timeout: 10,           # seconds
  idle_timeout: 120,           # seconds
  max_header_size: 1024 * 1024,  # 1MB
  max_body_size: 10 * 1024 * 1024  # 10MB
})
```

### Starting the Server

```ruby
# HTTP
app.listen(":8080")!
app.listen("127.0.0.1:8080")!

# HTTPS/TLS
app.listen_tls(":443", {
  cert: "./cert.pem",
  key: "./key.pem"
})!
```

### Graceful Shutdown

```ruby
app.on_shutdown do
  puts "shutting down..."
  db.close()!
  cache.close()!
end

app.listen(":8080")!
```

The server handles SIGINT and SIGTERM for graceful shutdown, waiting for active requests to complete.

## Complete Example

```ruby
# main.rg
import rugby/http_server
import rugby/log
import rugby/env
import "./routes/users"
import "./routes/posts"
import "./middleware/auth"
import "./db"

def main() -> error
  # Initialize database
  database = db.connect(env.get("DATABASE_URL")!)!
  defer database.close()

  # Create server
  app = http_server.new({
    read_timeout: 30,
    max_body_size: 5 * 1024 * 1024
  })

  # Global middleware
  app.use http_server.logger()
  app.use http_server.recover()
  app.use http_server.request_id()
  app.use http_server.compress()
  app.use http_server.cors({
    origins: [env.get("ALLOWED_ORIGIN") ?? "*"]
  })

  # Health check
  app.get "/health" do |ctx|
    ctx.json({ status: :ok })
  end

  # API routes
  app.mount "/api/v1" do |api|
    # Public routes
    api.post "/auth/login", auth.login_handler(database)
    api.post "/auth/register", auth.register_handler(database)

    # Protected routes
    api.mount "/users" do |users|
      users.use auth.require_auth()
      users.get "/", users.list_handler(database)
      users.get "/:id", users.get_handler(database)
      users.put "/:id", users.update_handler(database)
    end

    api.mount "/posts" do |posts|
      posts.use auth.optional_auth()
      posts.get "/", posts.list_handler(database)
      posts.get "/:id", posts.get_handler(database)

      posts.post "/", auth.require_auth(), posts.create_handler(database)
      posts.put "/:id", auth.require_auth(), posts.update_handler(database)
      posts.delete "/:id", auth.require_auth(), posts.delete_handler(database)
    end

    # Admin routes
    api.mount "/admin" do |admin|
      admin.use auth.require_auth()
      admin.use auth.require_role(:admin)
      admin.get "/stats", admin_stats_handler(database)
      admin.get "/users", admin_users_handler(database)
    end
  end

  # Error handlers
  app.on_error 404 do |ctx|
    ctx.json({ error: "not found" })
  end

  app.on_error 500 do |ctx, err|
    log.error("internal error", { error: err, path: ctx.path })
    ctx.json({ error: "internal server error" })
  end

  # Graceful shutdown
  app.on_shutdown do
    log.info("shutting down gracefully")
  end

  # Start server
  port = env.get("PORT") ?? "8080"
  log.info("starting server", { port: port })
  app.listen(":#{port}")!

  nil
end
```

## Types Reference

### http_server.App

The main application type.

| Method | Description |
|--------|-------------|
| `get(path, ...middleware) do \|ctx\| end` | Register GET handler |
| `post(path, ...middleware) do \|ctx\| end` | Register POST handler |
| `put(path, ...middleware) do \|ctx\| end` | Register PUT handler |
| `patch(path, ...middleware) do \|ctx\| end` | Register PATCH handler |
| `delete(path, ...middleware) do \|ctx\| end` | Register DELETE handler |
| `head(path, ...middleware) do \|ctx\| end` | Register HEAD handler |
| `options(path, ...middleware) do \|ctx\| end` | Register OPTIONS handler |
| `use(middleware)` | Add global middleware |
| `mount(prefix, routes)` | Mount routes at prefix |
| `mount(prefix) do \|r\| end` | Mount with block |
| `on_error(code?) do \|ctx, err?\| end` | Register error handler |
| `on_shutdown do end` | Register shutdown hook |
| `listen(addr)!` | Start HTTP server |
| `listen_tls(addr, options)!` | Start HTTPS server |

### http_server.Context

Request/response context passed to handlers.

| Method | Type | Description |
|--------|------|-------------|
| `method` | `String` | HTTP method |
| `path` | `String` | Request path |
| `param(name)` | `String` | Path parameter |
| `query(name)` | `String?` | Query parameter |
| `queries(name)` | `Array[String]` | Multi-value query param |
| `header(name)` | `String?` | Request header |
| `headers` | `Map[String, String]` | All headers |
| `ip` | `String` | Client IP |
| `json()` | `(Map[String, any], error)` | Parse JSON body |
| `json(Type)` | `(Type, error)` | Parse typed JSON |
| `text()` | `(String, error)` | Raw body text |
| `form()` | `(Map[String, String], error)` | Form data |
| `form_file(name)` | `(File, error)` | File upload |
| `cookie(name)` | `String?` | Cookie value |
| `status(code)` | | Set status |
| `status_code` | `Int` | Current status |
| `set_header(name, value)` | | Set response header |
| `set_cookie(name, value, options?)` | | Set cookie |
| `delete_cookie(name)` | | Delete cookie |
| `json(data)` | | Send JSON response |
| `text(data)` | | Send text response |
| `html(data)` | | Send HTML response |
| `file(path)` | `error` | Send file |
| `redirect(url, code?)` | | Redirect |
| `stream do \|w\| end` | | Streaming response |
| `set(key, value)` | | Set context value |
| `get(key)` | `any?` | Get context value |

### http_server.Routes

Route group for organizing handlers.

| Method | Description |
|--------|-------------|
| `get(path, ...middleware) do \|ctx\| end` | Register GET handler |
| `post(path, ...middleware) do \|ctx\| end` | Register POST handler |
| `put(path, ...middleware) do \|ctx\| end` | Register PUT handler |
| `patch(path, ...middleware) do \|ctx\| end` | Register PATCH handler |
| `delete(path, ...middleware) do \|ctx\| end` | Register DELETE handler |
| `use(middleware)` | Add middleware to group |
| `mount(prefix, routes)` | Nest routes |

### http_server.Middleware

Middleware function type. Create with `http_server.middleware`:

```ruby
http_server.middleware do |ctx, next_fn|
  # before handler
  next_fn.call()
  # after handler
end
```

### http_server.File

Uploaded file from multipart form.

| Field | Type | Description |
|-------|------|-------------|
| `filename` | `String` | Original filename |
| `size` | `Int` | Size in bytes |
| `content` | `Bytes` | File contents |
| `content_type` | `String` | MIME type |
| `save(path)` | `error` | Save to disk |
