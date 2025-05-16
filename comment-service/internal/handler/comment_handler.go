package handler

import (
	"net/http"
	"strconv"

	"github.com/Thedrogon/blogbish/comment-service/internal/models"
	"github.com/Thedrogon/blogbish/comment-service/internal/service"
	ws "github.com/Thedrogon/blogbish/comment-service/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type CommentHandler struct {
	commentService *service.CommentService
	hub            *ws.Hub
	upgrader       websocket.Upgrader
}

func NewCommentHandler(commentService *service.CommentService, hub *ws.Hub) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		hub:            hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // In production, implement proper origin checking
			},
		},
	}
}

func (h *CommentHandler) CreateComment(c *gin.Context) {
	var input models.CommentCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	input.UserID = userID.(int64)

	comment, err := h.commentService.CreateComment(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

func (h *CommentHandler) GetComment(c *gin.Context) {
	id := c.Param("id")

	comment, err := h.commentService.GetComment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	c.JSON(http.StatusOK, comment)
}

func (h *CommentHandler) UpdateComment(c *gin.Context) {
	id := c.Param("id")

	var input models.CommentUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	comment, err := h.commentService.UpdateComment(c.Request.Context(), id, userID.(int64), &input)
	if err != nil {
		switch err {
		case service.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		case service.ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to update this comment"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, comment)
}

func (h *CommentHandler) DeleteComment(c *gin.Context) {
	id := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.commentService.DeleteComment(c.Request.Context(), id, userID.(int64)); err != nil {
		switch err {
		case service.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		case service.ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to delete this comment"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CommentHandler) ListComments(c *gin.Context) {
	filter := &models.CommentFilter{
		PostID: c.Query("post_id"),
		Status: c.Query("status"),
	}

	if userID := c.Query("user_id"); userID != "" {
		id, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
			return
		}
		filter.UserID = id
	}

	if parentID := c.Query("parent_id"); parentID != "" {
		filter.ParentID = parentID
	}

	if page := c.Query("page"); page != "" {
		p, err := strconv.Atoi(page)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
			return
		}
		filter.Page = p
	}

	if pageSize := c.Query("page_size"); pageSize != "" {
		ps, err := strconv.Atoi(pageSize)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page size"})
			return
		}
		filter.PageSize = ps
	}

	comments, err := h.commentService.ListComments(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comments)
}

func (h *CommentHandler) LikeComment(c *gin.Context) {
	id := c.Param("id")

	if err := h.commentService.LikeComment(c.Request.Context(), id); err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *CommentHandler) ReportComment(c *gin.Context) {
	id := c.Param("id")

	if err := h.commentService.ReportComment(c.Request.Context(), id); err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *CommentHandler) ModerateComment(c *gin.Context) {
	id := c.Param("id")
	status := c.Query("status")

	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	if err := h.commentService.ModerateComment(c.Request.Context(), id, status); err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *CommentHandler) WebSocket(c *gin.Context) {
	postID := c.Query("post_id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post_id is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	client := ws.NewClient(h.hub, conn, postID, userID.(int64))
	h.hub.Register <- client

	// Start client message pumps
	go client.WritePump()
	go client.ReadPump()
}
