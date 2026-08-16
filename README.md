# Blog System

A full-featured blog platform built with Go, featuring JWT authentication, role-based access control, and a modern HTMX frontend.

## Features

### 🔐 Authentication
- JWT-based authentication with access (15 min) and refresh (7 days) tokens
- HTTP-only cookies for secure token storage
- Bcrypt password hashing
- Login, registration, and logout functionality

### 👥 User Management
- Admin dashboard for user management
- Role-based access control (Admin, Supervisor, User)
- Create, edit, toggle (enable/disable), and delete users
- User status management (active, disabled, banned)

### 📝 Post Management
- Create, edit, and delete posts
- Toggle post visibility (published/disabled)
- Pin important posts to the top
- Author attribution with timestamps
- Role-based permissions:
  - **Admin**: Full control over all posts
  - **Supervisor**: Full control over all posts
  - **User**: Manage only their own posts

### 🎨 Modern UI/UX
- HTMX for smooth, dynamic content updates without page reloads
- Templ for type-safe HTML templates
- Tailwind CSS for responsive, modern styling
- Clean, intuitive navigation
- Sidebar with quick links
- Latest posts displayed on homepage

### 💾 Database
- SQLite with proper migrations
- Clean repository pattern for database operations
- Automatic database initialization with default admin user

## Technology Stack

| Category | Technology |
|----------|------------|
| Backend | Go (standard library) |
| Authentication | JWT (golang-jwt/jwt) |
| Database | SQLite (mattn/go-sqlite3) |
| Frontend | HTMX, Tailwind CSS |
| Templates | Templ |
| Password Hashing | Bcrypt |

## Prerequisites

- Go 1.22 or higher
- SQLite3 (for database operations)
- Templ (for template generation)

## Installation

### 1. Clone the repository

```bash
git clone <your-repo-url>
cd blog-system
```
```


### 2. Install dependencies

```bash
go mod download
```

### 3. Install Templ

```bash
go install github.com/a-h/templ/cmd/templ@latest
```

### 4. Set up environment variables

Create a `.env` file in the root directory:

```env
JWT_SECRET=your-super-secret-key-change-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d
PORT=8080
```

### 5. Initialize the database

The database will be automatically initialized when you run the server. Default admin user will be created:

- Username: `admin`
- Password: `admin123`

### 6. Run migrations (for pinned posts feature)

```bash
go run run_migration.go
```

### 7. Generate Templ templates

```bash
templ generate ./...
```

### 8. Run the server

```bash
go run cmd/server/main.go
```

### 9. Access the application

Open your browser and navigate to: `http://localhost:8080`

## Project Structure

```
blog-system/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── auth/
│   │   ├── context.go              # User context handling
│   │   └── jwt.go                  # JWT token management
│   ├── handlers/
│   │   ├── auth.go                 # Authentication handlers
│   │   ├── post.go                 # Post CRUD handlers
│   │   └── user.go                 # User management handlers
│   ├── models/
│   │   └── models.go               # Data models
│   ├── repository/
│   │   ├── db.go                   # Database connection
│   │   ├── post.go                 # Post repository
│   │   └── user.go                 # User repository
│   ├── service/
│   │   ├── auth.go                 # Authentication service
│   │   ├── post.go                 # Post service
│   │   └── user.go                 # User service
│   └── views/
│       ├── components/             # Reusable Templ components
│       ├── layouts/                # Layout templates
│       └── pages/                  # Page templates
├── migrations/
│   ├── 001_initial.sql             # Initial schema
│   └── 002_add_pinned_to_posts.sql # Pinned posts feature
├── data/
│   └── blog.db                     # SQLite database (auto-created)
├── views/                          # Generated Templ files
├── .env                            # Environment variables
├── go.mod                          # Go module file
├── go.sum                          # Go dependencies
└── README.md                       # This file
```

## Usage Guide

### Default Admin Credentials

| Username | Password | Role |
| ------------- | -------------- | -------------- |
| admin | admin123 | Admin |

### User Roles

| Roles | Permissions |
| ------------- | -------------- |
| Admin | Full system access - manage users, all posts, roles |
| Admin | Manage all posts, view users (cannot modify users) |
| Admin | Manage only their own posts |

### Common Operations

For All Users

1. Register: Click "Register" in the navigation
2. Login: Click "Login" in the navigation
3. View Posts: Browse all published posts
4. Create Post: Click "New Post" (when logged in)

For Post Authors

1. Edit Post: Click "Edit" on your own post
2. Toggle Post: Click "Toggle" to enable/disable your post
3. Delete Post: Click "Delete" to remove your post

For Admin Users

1. Manage Users: Click "Users" in the navigation
2. Edit User: Click "Edit" on any user
3. Toggle User: Enable/disable user accounts
4. Delete User: Remove users (except yourself)
5. Change User Roles: Assign admin, supervisor, or user roles

Pinning Posts

To pin a post to the top of the homepage, run this SQL command:

```sql
UPDATE posts SET pinned = 1 WHERE id = [post_id];
```

## Development

### Generate Templ Templates

After making changes to `.templ` files:

```bash
templ generate ./...
```


### Run in Development Mode

```bash
go run cmd/server/main.go
```


### Build for Production

```bash
go build -o bin/blog-server cmd/server/main.go
./bin/blog-server
```


## API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/` | Home page | No |
| GET | `/login` | Login page | No |
| POST | `/login` | Login | No |
| GET | `/register` | Register page | No |
| POST | `/register` | Register | No |
| POST | `/logout` | Logout | No |
| GET | `/posts` | List posts | No |
| GET | `/posts/new` | Create post form | Yes |
| GET | `/posts/{id}/edit` | Edit post form | Yes |
| POST | `/posts/save` | Save post | Yes |
| DELETE | `/posts/{id}` | Delete post | Yes |
| POST | `/posts/{id}/toggle` | Toggle post status | Yes |
| GET | `/admin/users` | User management | Admin only |
| GET | `/admin/users/{id}/edit` | Edit user form | Admin only |
| POST | `/admin/users/update` | Update user | Admin only |
| DELETE | `/admin/users/{id}` | Delete user | Admin only |
| POST | `/admin/users/{id}/toggle-status` | Toggle user status | Admin only |


## Security Features

- JWT tokens with short-lived access tokens (15 minutes)
- HTTP-only cookies prevent XSS attacks
- Bcrypt password hashing
- Role-based access control
- Input validation on all forms
- CSRF protection via SameSite cookies
- Secure token refresh mechanism

## Troubleshooting

### Common Issues

1. **"templ: command not found"**
   ```bash
   go install github.com/a-h/templ/cmd/templ@latest
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

2. **"sqlite3: command not found"**
   ```bash
   # Ubuntu/Debian
   sudo apt-get install sqlite3
   # macOS
   brew install sqlite3
   ```

3. **Database locked error**
   ```bash
   # Stop the server and delete the lock file
   rm data/blog.db-journal
   ```

4. **Port already in use**
   ```bash
   # Change PORT in .env file or kill the process
   lsof -i :8080
   kill -9 [PID]
   ```

5. **"JWT_SECRET not set"**
   - Create a `.env` file with `JWT_SECRET=your-secret-key`

### Debugging

Enable debug logging in the application:

```go
// Add to main.go
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

Check the server logs for detailed error messages.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is open source and available under the MIT License.

## Acknowledgments

- [HTMX](https://htmx.org/) - Dynamic HTML without JavaScript
- [Templ](https://templ.guide/) - Type-safe HTML templating
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
- [SQLite](https://www.sqlite.org/) - Lightweight database

---

## Quick Start

```bash
# Clone and setup
git clone <your-repo-url>
cd blog-system
go mod download
go install github.com/a-h/templ/cmd/templ@latest

# Configure
cp .env.example .env  # Edit with your settings

# Run
templ generate ./...
go run cmd/server/main.go

# Access
# Open http://localhost:8080
# Login: admin / admin123
```

## Docker

### Build the image

```bash
# Build the Docker image
docker build -t blog-system .

# Or with a specific tag
docker build -t blog-system:latest .
```


### Run the container

```bash
# Run with environment variables
docker run -p 8080:8080 \
  -e JWT_SECRET=your-secret-key \
  -v $(pwd)/data:/app/data \
  blog-system

# Run with docker-compose
docker-compose up -d
```

## Manage Blog System Docker container

```bash
# List running containers
docker ps

# View logs
docker logs blog-system

# Stop the container
docker stop blog-system

# Remove the container
docker rm blog-system

# Remove the image
docker rmi blog-system

# Build without cache
docker build --no-cache -t blog-system .
```

## Support

For issues and questions, please open an issue on GitHub or contact the maintainers.

---

**Built with ❤️ using Go**
