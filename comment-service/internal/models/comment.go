package models

import (
	"time"
)

type Comment struct {
	ID        string     `json:"id"`
	PostID    string     `json:"post_id"`
	UserID    int64      `json:"user_id"`
	ParentID  *string    `json:"parent_id,omitempty"` // For nested comments
	Content   string     `json:"content"`
	Status    string     `json:"status"` // active, deleted, flagged, hidden
	Metadata  Metadata   `json:"metadata"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Children  []*Comment `json:"children,omitempty"`
}

type Metadata struct {
	Likes    int64      `json:"likes"`
	Reports  int64      `json:"reports"`
	EditedAt *time.Time `json:"edited_at,omitempty"`
	EditorID *int64     `json:"editor_id,omitempty"`
}

type CommentCreate struct {
	PostID   string `json:"post_id" binding:"required"`
	UserID   int64  `json:"user_id" binding:"required"`
	ParentID string `json:"parent_id,omitempty"`
	Content  string `json:"content" binding:"required,min=1,max=5000"`
}

type CommentUpdate struct {
	Content string `json:"content" binding:"required,min=1,max=5000"`
}

type CommentResponse struct {
	ID        string             `json:"id"`
	PostID    string             `json:"post_id"`
	UserID    int64              `json:"user_id"`
	ParentID  *string            `json:"parent_id,omitempty"`
	Content   string             `json:"content"`
	Status    string             `json:"status"`
	Metadata  Metadata           `json:"metadata"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	Children  []*CommentResponse `json:"children,omitempty"`
}

type CommentFilter struct {
	PostID   string
	UserID   int64
	ParentID string
	Status   string
	Page     int
	PageSize int
}

func (c *Comment) ToResponse() *CommentResponse {
	response := &CommentResponse{
		ID:        c.ID,
		PostID:    c.PostID,
		UserID:    c.UserID,
		ParentID:  c.ParentID,
		Content:   c.Content,
		Status:    c.Status,
		Metadata:  c.Metadata,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}

	if len(c.Children) > 0 {
		response.Children = make([]*CommentResponse, len(c.Children))
		for i, child := range c.Children {
			response.Children[i] = child.ToResponse()
		}
	}

	return response
}

// WebSocket event types
const (
	EventCommentCreated = "comment.created"
	EventCommentUpdated = "comment.updated"
	EventCommentDeleted = "comment.deleted"
	EventCommentLiked   = "comment.liked"
	EventCommentFlagged = "comment.flagged"
)

type WebSocketEvent struct {
	Type    string      `json:"type"`
	PostID  string      `json:"post_id"`
	Payload interface{} `json:"payload"`
}
