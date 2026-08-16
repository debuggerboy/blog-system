package main

import (
	"blog-system/internal/auth"
	"blog-system/internal/handlers"
	"blog-system/internal/models"
	"blog-system/internal/repository"
	"blog-system/internal/service"
	"blog-system/views/pages"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database
	if err := repository.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repository.CloseDB()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler()
	userHandler := handlers.NewUserHandler()
	postHandler := handlers.NewPostHandler()
	jwtService := auth.NewJWTService()

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Auth routes - these should NOT have the auth middleware
	mux.HandleFunc("GET /login", authHandler.LoginPage)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("GET /register", authHandler.RegisterPage)
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /logout", authHandler.Logout)
	mux.HandleFunc("POST /refresh-token", authHandler.RefreshToken)

	// Home route - public
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		user := auth.GetUserFromContext(r)

		// Get 5 latest posts
		postService := service.NewPostService()
		latestPosts, err := postService.GetLatestPosts(5)
		if err != nil {
			// Log error but continue with empty list
			log.Printf("Error fetching latest posts: %v", err)
			latestPosts = []models.Post{}
		}

		component := pages.Home(user, latestPosts)
		component.Render(r.Context(), w)
	})

	// Posts routes - public for viewing, protected for write operations
	mux.HandleFunc("GET /posts", postHandler.ListPosts)
	mux.HandleFunc("GET /posts/new", authHandler.AuthMiddleware(postHandler.ShowCreateForm))
	mux.HandleFunc("GET /posts/{id}/edit", authHandler.AuthMiddleware(postHandler.ShowEditForm))
	mux.HandleFunc("POST /posts/save", authHandler.AuthMiddleware(postHandler.SavePost))
	mux.HandleFunc("DELETE /posts/{id}", authHandler.AuthMiddleware(postHandler.DeletePost))
	mux.HandleFunc("POST /posts/{id}/toggle", authHandler.AuthMiddleware(postHandler.TogglePostStatus))

	// User routes (protected - admin only)
	mux.HandleFunc("GET /admin/users", authHandler.AuthMiddleware(userHandler.ListUsers))
	mux.HandleFunc("GET /admin/users/{id}/edit", authHandler.AuthMiddleware(userHandler.EditUser))
	mux.HandleFunc("POST /admin/users/update", authHandler.AuthMiddleware(userHandler.UpdateUser))
	mux.HandleFunc("DELETE /admin/users/{id}", authHandler.AuthMiddleware(userHandler.DeleteUser))
	mux.HandleFunc("POST /admin/users/{id}/toggle-status", authHandler.AuthMiddleware(userHandler.ToggleUserStatus))

	// Middleware for all routes to add user context from cookies (but NOT to intercept requests)
	handler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get user from access token cookie
			cookie, err := r.Cookie("access_token")
			if err == nil {
				claims, err := jwtService.ValidateToken(cookie.Value)
				if err == nil {
					userCtx := &auth.UserContext{
						UserID:   claims.UserID,
						Username: claims.Username,
						Email:    claims.Email,
						Role:     claims.Role,
					}
					ctx := auth.SetUserInContext(r.Context(), userCtx)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}

	// Wrap the mux with the context middleware
	http.Handle("/", handler(mux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
