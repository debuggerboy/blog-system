package handlers

import (
	"blog-system/internal/auth"
	"blog-system/internal/models"
	"blog-system/internal/service"
	"blog-system/views/pages"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type AuthHandler struct {
	authService *service.AuthService
	jwtService  *auth.JWTService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
		jwtService:  auth.NewJWTService(),
	}
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	// If already logged in, redirect to home
	if user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	component := pages.Login(nil)
	component.Render(r.Context(), w)
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	component := pages.Register(nil)
	component.Render(r.Context(), w)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	req.Username = r.FormValue("username")
	req.Password = r.FormValue("password")

	log.Printf("Login attempt: username=%s", req.Username)

	// Check if this is a JSON request
	isJSON := strings.Contains(r.Header.Get("Accept"), "application/json")

	tokens, user, err := h.authService.Login(&req)
	if err != nil {
		log.Printf("Login error: %v", err)
		if isJSON {
			writeJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Check if HTMX request
		if r.Header.Get("HX-Request") == "true" {
			// For HTMX, return only the login form with error message
			component := pages.LoginFormOnly("Invalid username or password")
			component.Render(r.Context(), w)
			return
		}

		// Full page for direct access
		component := pages.Login(nil)
		component.Render(r.Context(), w)
		return
	}

	log.Printf("Login successful for user: %s, role: %s", user.Username, user.Role)

	if isJSON {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokens)
		return
	}

	// Set both tokens as cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Path:     "/",
		MaxAge:   900,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		MaxAge:   604800,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})

	// For HTMX, redirect to home
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	req := &models.RegisterRequest{
		Username: r.FormValue("username"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	user, err := h.authService.Register(req)
	_ = user

	if err != nil {
		// Check if HTMX request
		if r.Header.Get("HX-Request") == "true" {
			// For HTMX, return only the register form with error message
			component := pages.RegisterFormOnly(err.Error())
			component.Render(r.Context(), w)
			return
		}
		// Full page for direct access
		component := pages.Register(nil)
		component.Render(r.Context(), w)
		return
	}

	// Auto-login after registration
	tokens, _, err := h.authService.Login(&models.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		if r.Header.Get("HX-Request") == "true" {
			component := pages.RegisterFormOnly("Registration successful but login failed. Please try logging in.")
			component.Render(r.Context(), w)
			return
		}
		component := pages.Register(nil)
		component.Render(r.Context(), w)
		return
	}

	// Set cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Path:     "/",
		MaxAge:   900,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		MaxAge:   604800,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})

	// Redirect to home
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		// Delete refresh token from database
		h.authService.Logout(cookie.Value)
	}

	// Clear cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// For HTMX, redirect to login with a full page refresh
	w.Header().Set("HX-Redirect", "/login")
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeJSONError(w, "Refresh token required", http.StatusUnauthorized)
		return
	}

	tokens, err := h.authService.RefreshToken(cookie.Value)
	if err != nil {
		writeJSONError(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Update cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Path:     "/",
		MaxAge:   900,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		MaxAge:   604800,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

// Helper function to write JSON errors
func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Middleware to validate JWT from cookie
func (h *AuthHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Try to get token from cookie first
		var tokenString string
		cookie, err := r.Cookie("access_token")
		if err == nil {
			tokenString = cookie.Value
		} else {
			// Fallback to Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}
		}

		if tokenString == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		claims, err := h.jwtService.ValidateToken(tokenString)
		if err != nil {
			// Try to refresh token
			refreshCookie, err := r.Cookie("refresh_token")
			if err == nil {
				// Attempt to refresh
				tokens, refreshErr := h.authService.RefreshToken(refreshCookie.Value)
				if refreshErr == nil {
					// Update cookies
					http.SetCookie(w, &http.Cookie{
						Name:     "access_token",
						Value:    tokens.AccessToken,
						Path:     "/",
						MaxAge:   900,
						HttpOnly: true,
						Secure:   false,
						SameSite: http.SameSiteStrictMode,
					})
					http.SetCookie(w, &http.Cookie{
						Name:     "refresh_token",
						Value:    tokens.RefreshToken,
						Path:     "/",
						MaxAge:   604800,
						HttpOnly: true,
						Secure:   false,
						SameSite: http.SameSiteStrictMode,
					})

					// Validate the new token
					claims, refreshErr = h.jwtService.ValidateToken(tokens.AccessToken)
					if refreshErr == nil {
						// Add user to context
						userCtx := &auth.UserContext{
							UserID:   claims.UserID,
							Username: claims.Username,
							Email:    claims.Email,
							Role:     claims.Role,
						}
						ctx := auth.SetUserInContext(r.Context(), userCtx)
						next(w, r.WithContext(ctx))
						return
					}
				}
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user to context
		userCtx := &auth.UserContext{
			UserID:   claims.UserID,
			Username: claims.Username,
			Email:    claims.Email,
			Role:     claims.Role,
		}
		ctx := auth.SetUserInContext(r.Context(), userCtx)
		next(w, r.WithContext(ctx))
	}
}
