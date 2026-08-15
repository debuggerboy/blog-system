package handlers

import (
	"blog-system/internal/auth"
	"blog-system/internal/models"
	"blog-system/internal/service"
	"blog-system/views/components"
	"blog-system/views/pages"
	"strconv"

	"net/http"
)

type PostHandler struct {
	postService *service.PostService
}

func NewPostHandler() *PostHandler {
	return &PostHandler{
		postService: service.NewPostService(),
	}
}

func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)

	limit := 20
	offset := 0

	// Get pagination params
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var posts []models.Post
	var err error

	if user != nil {
		posts, err = h.postService.ListPosts(limit, offset, user.UserID, user.Role)
	} else {
		posts, err = h.postService.ListPosts(limit, offset, 0, "")
	}

	if err != nil {
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	// Check if HTMX request
	if r.Header.Get("HX-Request") == "true" {
		component := components.PostList(posts, user)
		component.Render(r.Context(), w)
		return
	}

	component := pages.PostListPage(posts, user)
	component.Render(r.Context(), w)
}

func (h *PostHandler) ShowCreateForm(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	component := components.PostForm(nil, user)
	component.Render(r.Context(), w)
}

func (h *PostHandler) ShowEditForm(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Post ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := h.postService.GetPostByID(id)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Check permissions
	if user.Role != "admin" && user.Role != "supervisor" && post.AuthorID != user.UserID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	component := components.PostForm(post, user)
	component.Render(r.Context(), w)
}

func (h *PostHandler) SavePost(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	req := &models.PostRequest{
		Title:   r.FormValue("title"),
		Content: r.FormValue("content"),
		Status:  r.FormValue("status"),
	}

	if req.Status == "" {
		req.Status = "published"
	}

	idStr := r.FormValue("id")
	var post *models.Post
	var err error

	if idStr != "" {
		// Update existing post
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}
		post, err = h.postService.UpdatePost(id, req, user.UserID, user.Role)
	} else {
		// Create new post
		post, err = h.postService.CreatePost(req, user.UserID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_ = post // Mark as used

	// Redirect to post list
	w.Header().Set("HX-Redirect", "/posts")
	w.WriteHeader(http.StatusOK)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Post ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	err = h.postService.DeletePost(id, user.UserID, user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PostHandler) TogglePostStatus(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Post ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := h.postService.TogglePostStatus(id, user.UserID, user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return updated post card
	component := components.PostCard(*post, user)
	component.Render(r.Context(), w)
}
