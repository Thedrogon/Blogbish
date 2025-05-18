package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Thedrogon/blogbish/post-service/internal/models"
	"github.com/Thedrogon/blogbish/post-service/internal/service"
	"github.com/go-chi/chi/v5"
)

type CategoryHandler struct {
	categoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.CategoryCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	category, err := h.categoryService.CreateCategory(r.Context(), &input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.Error(w, "Missing category slug", http.StatusBadRequest)
		return
	}

	category, err := h.categoryService.GetCategory(r.Context(), slug)
	if err != nil {
		if err == service.ErrNotFound {
			http.Error(w, "Category not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.Error(w, "Missing category slug", http.StatusBadRequest)
		return
	}

	var input models.CategoryUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	category, err := h.categoryService.UpdateCategory(r.Context(), slug, &input)
	if err != nil {
		if err == service.ErrNotFound {
			http.Error(w, "Category not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.Error(w, "Missing category slug", http.StatusBadRequest)
		return
	}

	if err := h.categoryService.DeleteCategory(r.Context(), slug); err != nil {
		if err == service.ErrNotFound {
			http.Error(w, "Category not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryService.ListCategories(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}
