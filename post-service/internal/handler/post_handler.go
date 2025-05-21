package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Thedrogon/blogbish/post-service/internal/models"
	"github.com/Thedrogon/blogbish/post-service/internal/service"
	"github.com/go-chi/chi/v5"
)

type PostHandler struct {
	postService service.PostService
}

func NewPostHandler(postService service.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,

		
	}
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	var authorID int64
	var input models.PostCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	post, err := h.postService.CreatePost(r.Context(), &input , authorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func (h *PostHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	post, err := h.postService.GetPost(r.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	var input models.PostUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	post, err := h.postService.UpdatePost(r.Context(), id, &input)
	if err != nil {
		if err == service.ErrNotFound {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	if err := h.postService.DeletePost(r.Context(), id); err != nil {
		if err == service.ErrNotFound {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	categorySlug := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")
	tag := r.URL.Query().Get("tag")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filter := &models.PostFilter{
		Page:         page,
		PageSize:     pageSize,
		CategorySlug: categorySlug,
		Status:       status,
		Tag:         tag,
	}

	posts, err := h.postService.ListPosts(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
} 