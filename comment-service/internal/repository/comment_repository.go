package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/Thedrogon/blogbish/comment-service/internal/models"
)

// type CommentRepository interface {                 //Redeclared
// 	Create(ctx context.Context, comment *models.Comment) error
// 	GetByID(ctx context.Context, id string) (*models.Comment, error)
// 	Update(ctx context.Context, comment *models.Comment) error
// 	Delete(ctx context.Context, id string) error
// 	List(ctx context.Context, filter *models.CommentFilter) ([]*models.Comment, error)
// 	Like(ctx context.Context, id string) error
// 	Report(ctx context.Context, id string) error
// 	UpdateStatus(ctx context.Context, id string, status string) error
// }

type commentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *models.Comment) error {
	metadata, err := json.Marshal(comment.Metadata)
	if err != nil {
		return err
	}

	query := `INSERT INTO comments (id, post_id, user_id, content, parent_id, status, metadata) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = r.db.ExecContext(ctx, query, comment.ID, comment.PostID, comment.UserID,
		comment.Content, comment.ParentID, comment.Status, metadata)
	return err
}

func (r *commentRepository) GetByID(ctx context.Context, id string) (*models.Comment, error) {
	var comment models.Comment
	var metadata []byte

	query := `SELECT id, post_id, user_id, content, parent_id, status, metadata, created_at, updated_at 
			  FROM comments WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&comment.ID, &comment.PostID,
		&comment.UserID, &comment.Content, &comment.ParentID, &comment.Status,
		&metadata, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(metadata, &comment.Metadata); err != nil {
		return nil, err
	}

	// Get children if any
	children, err := r.getChildren(ctx, id)
	if err != nil {
		return nil, err
	}
	comment.Children = children

	return &comment, nil
}

func (r *commentRepository) Update(ctx context.Context, comment *models.Comment) error {
	metadata, err := json.Marshal(comment.Metadata)
	if err != nil {
		return err
	}

	query := `UPDATE comments SET content = $1, status = $2, metadata = $3, updated_at = NOW() 
			  WHERE id = $4 AND user_id = $5`
	_, err = r.db.ExecContext(ctx, query, comment.Content, comment.Status, metadata,
		comment.ID, comment.UserID)
	return err
}

func (r *commentRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE comments SET status = 'deleted', updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *commentRepository) List(ctx context.Context, filter *models.CommentFilter) ([]*models.Comment, error) {
	query := `SELECT id, post_id, user_id, content, parent_id, status, metadata, created_at, updated_at 
			  FROM comments WHERE post_id = $1`
	args := []interface{}{filter.PostID}

	if filter.ParentID != "" {
		query += ` AND parent_id = $2`
		args = append(args, filter.ParentID)
	}
	if filter.UserID != 0 {
		query += ` AND user_id = $3`
		args = append(args, filter.UserID)
	}
	if filter.Status != "" {
		query += ` AND status = $4`
		args = append(args, filter.Status)
	}

	query += ` ORDER BY created_at DESC`

	if filter.PageSize > 0 {
		query += ` LIMIT $5 OFFSET $6`
		args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment
		var metadata []byte
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content,
			&comment.ParentID, &comment.Status, &metadata, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadata, &comment.Metadata); err != nil {
			return nil, err
		}

		comments = append(comments, &comment)
	}

	// Get children for top-level comments
	if filter.ParentID == "" {
		for _, comment := range comments {
			children, err := r.getChildren(ctx, comment.ID)
			if err != nil {
				return nil, err
			}
			comment.Children = children
		}
	}

	return comments, nil
}

func (r *commentRepository) Like(ctx context.Context, id string) error {
	query := `UPDATE comments SET metadata = jsonb_set(metadata, '{likes}', 
			  (COALESCE(metadata->>'likes','0')::int + 1)::text::jsonb) WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *commentRepository) Report(ctx context.Context, id string) error {
	query := `UPDATE comments SET metadata = jsonb_set(metadata, '{reports}', 
			  (COALESCE(metadata->>'reports','0')::int + 1)::text::jsonb) WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *commentRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE comments SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *commentRepository) getChildren(ctx context.Context, parentID string) ([]*models.Comment, error) {
	query := `SELECT id, post_id, user_id, content, parent_id, status, metadata, created_at, updated_at 
			  FROM comments WHERE parent_id = $1 ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []*models.Comment
	for rows.Next() {
		var comment models.Comment
		var metadata []byte
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content,
			&comment.ParentID, &comment.Status, &metadata, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadata, &comment.Metadata); err != nil {
			return nil, err
		}

		children = append(children, &comment)
	}

	return children, nil
}
