package handler

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/Thedrogon/blogbish/media-service/internal/models"
	"github.com/Thedrogon/blogbish/media-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MediaHandler struct {
	mediaService *service.MediaService
}

func NewMediaHandler(mediaService *service.MediaService) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
	}
}

func (h *MediaHandler) Upload(c *gin.Context) {
	// Parse multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Create upload metadata
	upload := &models.MediaUpload{
		UserID:      userID.(int64),
		Title:       c.PostForm("title"),
		Description: c.PostForm("description"),
		AltText:     c.PostForm("alt_text"),
	}

	// Upload file
	media, err := h.mediaService.UploadFile(c.Request.Context(), file, filename, header.Header.Get("Content-Type"), upload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, media)
}

func (h *MediaHandler) Download(c *gin.Context) {
	id := c.Param("id")

	media, err := h.mediaService.GetFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	reader, err := h.mediaService.DownloadFile(c.Request.Context(), media.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file"})
		return
	}
	defer reader.Close()

	c.Header("Content-Type", media.ContentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", media.Filename))
	c.DataFromReader(http.StatusOK, media.Size, media.ContentType, reader, nil)
}

func (h *MediaHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Delete file
	if err := h.mediaService.DeleteFile(c.Request.Context(), id, userID.(int64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *MediaHandler) GetMetadata(c *gin.Context) {
	id := c.Param("id")

	media, err := h.mediaService.GetFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.JSON(http.StatusOK, media)
}

func (h *MediaHandler) UpdateMetadata(c *gin.Context) {
	id := c.Param("id")

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var metadata models.Metadata
	if err := c.ShouldBindJSON(&metadata); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metadata"})
		return
	}

	media, err := h.mediaService.UpdateMetadata(c.Request.Context(), id, userID.(int64), &metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, media)
}
