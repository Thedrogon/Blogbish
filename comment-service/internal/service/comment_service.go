package service

import (
	"context"
	"errors"
	"time"

	"github.com/Thedrogon/blogbish/comment-service/internal/models"
	"github.com/Thedrogon/blogbish/comment-service/internal/repository"
	"github.com/Thedrogon/blogbish/comment-service/internal/websocket"
	"github.com/google/uuid"
)

var (
	ErrNotFound  = errors.New("comment not found")
	ErrForbidden = errors.New("forbidden")
)

type CommentService struct {
	repo repository.CommentRepository
	hub  *websocket.Hub
}

func NewCommentService(repo repository.CommentRepository, hub *websocket.Hub) *CommentService {
	return &CommentService{
		repo: repo,
		hub:  hub,
	}
}

func (s *CommentService) CreateComment(ctx context.Context, input *models.CommentCreate) (*models.CommentResponse, error) {
	comment := &models.Comment{
		ID:        uuid.New().String(),
		PostID:    input.PostID,
		UserID:    input.UserID,
		Content:   input.Content,
		Status:    "active",
		Metadata:  models.Metadata{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if input.ParentID != "" {
		comment.ParentID = &input.ParentID
	}

	if err := s.repo.Create(ctx, comment); err != nil {
		return nil, err
	}

	// Broadcast comment creation event
	event := &models.WebSocketEvent{
		Type:    models.EventCommentCreated,
		PostID:  comment.PostID,
		Payload: comment.ToResponse(),
	}
	s.hub.BroadcastToPost(comment.PostID, event)

	return comment.ToResponse(), nil
}

func (s *CommentService) GetComment(ctx context.Context, id string) (*models.CommentResponse, error) {
	comment, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	return comment.ToResponse(), nil
}

func (s *CommentService) UpdateComment(ctx context.Context, id string, userID int64, input *models.CommentUpdate) (*models.CommentResponse, error) {
	comment, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	if comment.UserID != userID {
		return nil, ErrForbidden
	}

	comment.Content = input.Content
	comment.UpdatedAt = time.Now()
	comment.Metadata.EditedAt = &comment.UpdatedAt
	comment.Metadata.EditorID = &userID

	if err := s.repo.Update(ctx, comment); err != nil {
		return nil, err
	}

	// Broadcast comment update event
	event := &models.WebSocketEvent{
		Type:    models.EventCommentUpdated,
		PostID:  comment.PostID,
		Payload: comment.ToResponse(),
	}
	s.hub.BroadcastToPost(comment.PostID, event)

	return comment.ToResponse(), nil
}

func (s *CommentService) DeleteComment(ctx context.Context, id string, userID int64) error {
	comment, err := s.repo.Get(ctx, id)
	if err != nil {
		return ErrNotFound
	}

	if comment.UserID != userID {
		return ErrForbidden
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Broadcast comment deletion event
	event := &models.WebSocketEvent{
		Type:    models.EventCommentDeleted,
		PostID:  comment.PostID,
		Payload: map[string]string{"id": id},
	}
	s.hub.BroadcastToPost(comment.PostID, event)

	return nil
}

func (s *CommentService) ListComments(ctx context.Context, filter *models.CommentFilter) ([]*models.CommentResponse, error) {
	comments, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]*models.CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = comment.ToResponse()
	}

	return responses, nil
}

func (s *CommentService) LikeComment(ctx context.Context, id string) error {
	if err := s.repo.IncrementLikes(ctx, id); err != nil {
		return err
	}

	comment, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Broadcast like event
	event := &models.WebSocketEvent{
		Type:    models.EventCommentLiked,
		PostID:  comment.PostID,
		Payload: map[string]interface{}{"id": id, "likes": comment.Metadata.Likes},
	}
	s.hub.BroadcastToPost(comment.PostID, event)

	return nil
}

func (s *CommentService) ReportComment(ctx context.Context, id string) error {
	if err := s.repo.IncrementReports(ctx, id); err != nil {
		return err
	}

	comment, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// If reports exceed threshold, flag the comment for moderation
	if comment.Metadata.Reports >= 5 {
		comment.Status = "flagged"
		if err := s.repo.UpdateStatus(ctx, id, "flagged"); err != nil {
			return err
		}

		// Broadcast flagged event
		event := &models.WebSocketEvent{
			Type:    models.EventCommentFlagged,
			PostID:  comment.PostID,
			Payload: comment.ToResponse(),
		}
		s.hub.BroadcastToPost(comment.PostID, event)
	}

	return nil
}

func (s *CommentService) ModerateComment(ctx context.Context, id string, status string) error {
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	comment, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Broadcast update event
	event := &models.WebSocketEvent{
		Type:    models.EventCommentUpdated,
		PostID:  comment.PostID,
		Payload: comment.ToResponse(),
	}
	s.hub.BroadcastToPost(comment.PostID, event)

	return nil
}
