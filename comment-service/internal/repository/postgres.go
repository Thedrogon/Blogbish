package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Thedrogon/blogbish/comment-service/internal/models"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, comment *models.Comment) error {
	query := `
		INSERT INTO comments (
			id, post_id, user_id, parent_id, content, status,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	metadata, err := encodeMetadata(comment.Metadata)
	if err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		comment.ID,
		comment.PostID,
		comment.UserID,
		comment.ParentID,
		comment.Content,
		comment.Status,
		metadata,
		comment.CreatedAt,
		comment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (*models.Comment, error) {
	query := `
		SELECT id, post_id, user_id, parent_id, content, status,
			   metadata, created_at, updated_at
		FROM comments
		WHERE id = $1 AND status != 'deleted'
	`

	comment := &models.Comment{}
	var metadata []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.ParentID,
		&comment.Content,
		&comment.Status,
		&metadata,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("comment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	if err := decodeMetadata(metadata, &comment.Metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	// Get children comments
	children, err := r.GetChildren(ctx, comment.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	comment.Children = children

	return comment, nil
}

func (r *PostgresRepository) Update(ctx context.Context, comment *models.Comment) error {
	query := `
		UPDATE comments
		SET content = $1, status = $2, metadata = $3, updated_at = $4
		WHERE id = $5 AND status != 'deleted'
	`

	metadata, err := encodeMetadata(comment.Metadata)
	if err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		comment.Content,
		comment.Status,
		metadata,
		time.Now(),
		comment.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("comment not found")
	}

	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE comments
		SET status = 'deleted', updated_at = $1
		WHERE id = $2 OR parent_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}

func (r *PostgresRepository) List(ctx context.Context, filter *models.CommentFilter) ([]*models.Comment, error) {
	query := `
		SELECT id, post_id, user_id, parent_id, content, status,
			   metadata, created_at, updated_at
		FROM comments
		WHERE 1=1
	`
	var args []interface{}
	var conditions []string

	if filter.PostID != "" {
		args = append(args, filter.PostID)
		conditions = append(conditions, fmt.Sprintf("post_id = $%d", len(args)))
	}

	if filter.UserID != 0 {
		args = append(args, filter.UserID)
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", len(args)))
	}

	if filter.ParentID != "" {
		args = append(args, filter.ParentID)
		conditions = append(conditions, fmt.Sprintf("parent_id = $%d", len(args)))
	}

	if filter.Status != "" {
		args = append(args, filter.Status)
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)))
	} else {
		conditions = append(conditions, "status != 'deleted'")
	}

	for _, condition := range conditions {
		query += " AND " + condition
	}

	query += " ORDER BY created_at DESC"

	if filter.PageSize > 0 {
		args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		comment := &models.Comment{}
		var metadata []byte

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.ParentID,
			&comment.Content,
			&comment.Status,
			&metadata,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		if err := decodeMetadata(metadata, &comment.Metadata); err != nil {
			return nil, fmt.Errorf("failed to decode metadata: %w", err)
		}

		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *PostgresRepository) GetChildren(ctx context.Context, parentID string) ([]*models.Comment, error) {
	query := `
		SELECT id, post_id, user_id, parent_id, content, status,
			   metadata, created_at, updated_at
		FROM comments
		WHERE parent_id = $1 AND status != 'deleted'
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		comment := &models.Comment{}
		var metadata []byte

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.ParentID,
			&comment.Content,
			&comment.Status,
			&metadata,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan child comment: %w", err)
		}

		if err := decodeMetadata(metadata, &comment.Metadata); err != nil {
			return nil, fmt.Errorf("failed to decode metadata: %w", err)
		}

		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE comments
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update comment status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("comment not found")
	}

	return nil
}

func (r *PostgresRepository) IncrementLikes(ctx context.Context, id string) error {
	query := `
		UPDATE comments
		SET metadata = jsonb_set(
			metadata::jsonb,
			'{likes}',
			(COALESCE((metadata->>'likes')::int, 0) + 1)::text::jsonb
		)
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment likes: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("comment not found")
	}

	return nil
}

func (r *PostgresRepository) IncrementReports(ctx context.Context, id string) error {
	query := `
		UPDATE comments
		SET metadata = jsonb_set(
			metadata::jsonb,
			'{reports}',
			(COALESCE((metadata->>'reports')::int, 0) + 1)::text::jsonb
		)
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment reports: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("comment not found")
	}

	return nil
}

// Helper functions for metadata encoding/decoding
func encodeMetadata(metadata models.Metadata) ([]byte, error) {
	return json.Marshal(metadata)
}

func decodeMetadata(data []byte, metadata *models.Metadata) error {
	return json.Unmarshal(data, metadata)
}
