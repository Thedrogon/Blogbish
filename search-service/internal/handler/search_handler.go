package handler

import (
	"net/http"

	"github.com/Thedrogon/blogbish/search-service/internal/models"
	"github.com/Thedrogon/blogbish/search-service/internal/service"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

func (h *SearchHandler) Search(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default values if not provided
	if req.Size == 0 {
		req.Size = 10
	}
	if req.SortBy == "" {
		req.SortBy = "relevance"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	result, err := h.searchService.Search(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *SearchHandler) Suggest(c *gin.Context) {
	var req models.SuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default values if not provided
	if req.Limit == 0 {
		req.Limit = 5
	}

	result, err := h.searchService.Suggest(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *SearchHandler) IndexPost(c *gin.Context) {
	var post models.SearchablePost
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.searchService.IndexPost(c.Request.Context(), &post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

func (h *SearchHandler) IndexComment(c *gin.Context) {
	var comment models.SearchableComment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.searchService.IndexComment(c.Request.Context(), &comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}
