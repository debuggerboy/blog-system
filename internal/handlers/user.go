package handlers

import (
	"blog-system/internal/auth"
	"blog-system/internal/models"
	"blog-system/internal/service"
	"blog-system/views/pages"
	"strconv"

	"net/http"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(),
	}
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil || (user.Role != "admin" && user.Role != "supervisor") {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	users, err := h.userService.ListUsers()
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	// Check if HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return only the user list table, not the full layout
		component := pages.UserListOnly(users, user)
		component.Render(r.Context(), w)
		return
	}

	component := pages.UserList(users, user)
	component.Render(r.Context(), w)
}

func (h *UserHandler) EditUser(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil || user.Role != "admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get user ID from URL
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	targetUser, err := h.userService.GetUserByID(id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return only the form component, not the full layout
		component := pages.UserFormOnly(targetUser, user)
		component.Render(r.Context(), w)
		return
	}

	// Full page for direct access
	component := pages.UserEditForm(targetUser, user)
	component.Render(r.Context(), w)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	admin := auth.GetUserFromContext(r)
	if admin == nil || admin.Role != "admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	req := &models.UserUpdateRequest{
		Username: r.FormValue("username"),
		Email:    r.FormValue("email"),
		Role:     r.FormValue("role"),
		Status:   r.FormValue("status"),
	}

	_, err = h.userService.UpdateUser(id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If updating password
	if newPassword := r.FormValue("new_password"); newPassword != "" {
		err = h.userService.UpdatePassword(id, newPassword)
		if err != nil {
			http.Error(w, "Error updating password", http.StatusInternalServerError)
			return
		}
	}

	// Redirect to user list with HX-Refresh to force full page reload
	w.Header().Set("HX-Redirect", "/admin/users")
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	admin := auth.GetUserFromContext(r)
	if admin == nil || admin.Role != "admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Don't allow deleting yourself
	if id == admin.UserID {
		http.Error(w, "Cannot delete your own account", http.StatusBadRequest)
		return
	}

	err = h.userService.DeleteUser(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) ToggleUserStatus(w http.ResponseWriter, r *http.Request) {
	admin := auth.GetUserFromContext(r)
	if admin == nil || admin.Role != "admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Toggle status
	if user.Status == "active" {
		user.Status = "disabled"
	} else {
		user.Status = "active"
	}

	// Don't allow disabling the last admin
	if user.Role == "admin" && user.Status == "disabled" {
		users, _ := h.userService.ListUsers()
		adminCount := 0
		for _, u := range users {
			if u.Role == "admin" && u.Status == "active" {
				adminCount++
			}
		}
		if adminCount <= 1 {
			http.Error(w, "Cannot disable the last active admin", http.StatusBadRequest)
			return
		}
	}

	updatedUser, err := h.userService.UpdateUser(id, &models.UserUpdateRequest{
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Status:   user.Status,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated user row - pass the pointer directly
	component := pages.UserRow(*updatedUser, admin)
	component.Render(r.Context(), w)
}
