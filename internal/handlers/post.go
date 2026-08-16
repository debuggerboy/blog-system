package handlers

import (
	"blog-system/internal/auth"
	"blog-system/internal/models"
	"blog-system/internal/service"
	"blog-system/views/components"
	"blog-system/views/pages"
	"log"
	"net/http"
	"strconv"
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

	if r.Header.Get("HX-Request") == "true" {
		component := components.PostFormOnly(nil, user)
		component.Render(r.Context(), w)
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

	if user.Role != "admin" && user.Role != "supervisor" && post.AuthorID != user.UserID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		component := components.PostFormOnly(post, user)
		component.Render(r.Context(), w)
		return
	}

	component := components.PostForm(post, user)
	component.Render(r.Context(), w)
}

func (h *PostHandler) SavePost(w http.ResponseWriter, r *http.Request) {
	log.Println("SavePost called")

	user := auth.GetUserFromContext(r)
	if user == nil {
		log.Println("User not authenticated")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	status := r.FormValue("status")
	idStr := r.FormValue("id")

	log.Printf("Form values - Title: %s, Content: %s, Status: %s, ID: %s", title, content, status, idStr)

	req := &models.PostRequest{
		Title:   title,
		Content: content,
		Status:  status,
	}

	if req.Status == "" {
		req.Status = "published"
	}

	var post *models.Post
	var err error

	if idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Invalid post ID: %v", err)
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}
		log.Printf("Updating post ID: %d", id)
		post, err = h.postService.UpdatePost(id, req, user.UserID, user.Role)
	} else {
		log.Println("Creating new post")
		post, err = h.postService.CreatePost(req, user.UserID)
	}

	if err != nil {
		log.Printf("Error saving post: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Post saved successfully with ID: %d", post.ID)

	// Redirect to posts page with HX-Refresh to force full page reload
	w.Header().Set("HX-Redirect", "/posts")
	w.Header().Set("HX-Refresh", "true")
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

	component := components.PostCard(*post, user)
	component.Render(r.Context(), w)
}
